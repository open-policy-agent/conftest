// Package textproto provides an interface to parse Protocol Buffers in their
// textual format.
//
// https://protobuf.dev/reference/protobuf/textformat-spec/
package textproto

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"

	protoparser "github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

// TextProtoFileExtensions is the list of all file extensions associated with
// textproto files.
var TextProtoFileExtensions = []string{"textproto", "textpb"}

var (
	msgTypeRegexp = regexp.MustCompile(`#\s+proto-message:\s+([a-zA-Z0-9\.]+)`)
	marshaller    = protojson.MarshalOptions{
		UseProtoNames: true, // Keep field names 1-to-1 with proto field definitions.
	}
	globalFiles = protoregistry.GlobalFiles // Alias for convenience.
	globalTypes = protoregistry.GlobalTypes // Alias for convenience.
)

// Parser provides methods to parse textproto files.
type Parser struct{}

// LoadProtoFiles loads Protocol Buffer definitions so that the textproto files
// can be parsed correctly.
func (p *Parser) LoadProtoFiles(filePaths []string) error {
	if len(filePaths) == 0 {
		return fmt.Errorf("must supply at least one file path")
	}
	if err := loadFiles(filePaths); err != nil {
		return fmt.Errorf("load messages from files: %w", err)
	}
	return nil
}

// Unmarshal unmarshals a textproto file.
func (p *Parser) Unmarshal(data []byte, v any) error {
	ty, err := extractMsgType(data)
	if err != nil {
		return fmt.Errorf("extract proto message type: %w", err)
	}
	desc, err := globalTypes.FindMessageByName(protoreflect.FullName(ty))
	if err != nil {
		return fmt.Errorf("look up message type %q: %w", ty, err)
	}
	msg, ok := desc.Zero().(protoreflect.ProtoMessage)
	if !ok {
		return fmt.Errorf("could not assert ProtoMessage for %q", ty)
	}
	if err := prototext.Unmarshal(data, msg); err != nil {
		return fmt.Errorf("unmarshal textproto: %w", err)
	}

	by, err := marshaller.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	if err := json.Unmarshal(by, v); err != nil {
		return fmt.Errorf("unmarshal JSON: %w", err)
	}

	return nil
}

func loadFiles(paths []string) error {
	for _, path := range paths {
		if err := loadFile(path); err != nil {
			return fmt.Errorf("load proto file %s: %w", path, err)
		}
	}
	return nil
}

func loadFile(path string) error {
	fh, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer fh.Close()
	return load(path, fh)
}

func load(fileName string, r io.Reader) error {
	// Only load each file once.
	if _, err := globalFiles.FindFileByPath(fileName); err == nil {
		return nil
	}

	handler := reporter.NewHandler(nil)
	node, err := protoparser.Parse(fileName, r, handler)
	if err != nil {
		return fmt.Errorf("parse proto: %w", err)
	}
	res, err := protoparser.ResultFromAST(node, true /* validate */, handler)
	if err != nil {
		return fmt.Errorf("convert from AST: %w", err)
	}
	fd, err := protodesc.NewFile(res.FileDescriptorProto(), globalFiles)
	if err != nil {
		return fmt.Errorf("convert to FileDescriptor: %w", err)
	}
	if err := globalFiles.RegisterFile(fd); err != nil {
		return fmt.Errorf("register file: %w", err)
	}
	for i := 0; i < fd.Messages().Len(); i++ {
		msg := fd.Messages().Get(i)
		if err := globalTypes.RegisterMessage(dynamicpb.NewMessageType(msg)); err != nil {
			return fmt.Errorf("register message %q: %w", msg.FullName(), err)
		}
	}
	for i := 0; i < fd.Extensions().Len(); i++ {
		ext := fd.Extensions().Get(i)
		if err := globalTypes.RegisterExtension(dynamicpb.NewExtensionType(ext)); err != nil {
			return fmt.Errorf("register extension %q: %w", ext.FullName(), err)
		}
	}

	return nil
}

func extractMsgType(data []byte) (string, error) {
	s := bufio.NewScanner(bytes.NewReader(data))
	for s.Scan() {
		matches := msgTypeRegexp.FindStringSubmatch(s.Text())
		if len(matches) == 0 {
			continue
		}
		return matches[1], nil
	}

	return "", fmt.Errorf("file does not contain necessary proto-message comment")
}
