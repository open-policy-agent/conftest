package vcl

import "testing"

func TestVCLParser(t *testing.T) {
	parser := &Parser{}
	sample := `acl purge {
	"127.0.0.1";
	"localhost";
}


backend default {
    .host = "127.0.0.1";
    .port = "8080";
    .connect_timeout = 60s;
    .first_byte_timeout = 60s;
	.between_bytes_timeout = 60s;
	.max_connections = 800;
}

sub vcl_recv {
    set req.grace = 2m;

    remove req.http.X-Forwarded-For;
    set    req.http.X-Forwarded-For = client.ip;

    set req.http.Cookie = regsuball(req.http.Cookie, "(^|;\s*)(_[_a-z]+|has_js|is_
    unique)=[^;]*", "");
    set req.http.Cookie = regsub(req.http.Cookie, "^;\s*", "");

    if (req.url ~ "/wp-(login|admin|cron)") {
        return (pass);
    }

    set req.http.Cookie = regsuball(req.http.Cookie, "wp-settings-1=[^;]+(; )?", "")
    ;

    set req.http.Cookie = regsuball(req.http.Cookie, "wp-settings-time-1=[^;]+(; )?"
    , "");

    set req.http.Cookie = regsuball(req.http.Cookie, "wordpress_test_cookie=[^;]+(; 
    )?", "");

    if (req.url ~ "wp-content/themes/" && req.url ~ "\.(css|js|png|gif|jp(e)?g)") {
        unset req.http.cookie;
    }

    if (req.url ~ "/wp-content/uploads/") {
        return (pass);
    }

    if (req.url ~ "^/contact/" || req.url ~ "^/links/domains-for-sale/")
        {
            return(pass);
        }

    if (req.http.Cookie ~ "wordpress_" || req.http.Cookie ~ "comment_") {
        return (pass);
    }

    if (req.request == "PURGE") {
        if (!client.ip ~ purge) {
            error 405 "Not allowed.";
        }
        return (lookup);
    }

    if (req.http.Cache-Control ~ "no-cache") {
        return (pass);
    }

    return (lookup);
}
`

	var input interface{}
	if err := parser.Unmarshal([]byte(sample), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("there should be information parsed but its nil")
	}

	item := input.(map[string]interface{})

	if len(item) <= 0 {
		t.Error("there should be at least one item defined in the parsed file, but none found")
	}
}
