package terraform

import (
	"testing"
)

func TestTerraformParser(t *testing.T) {
	parser := &Parser{
		FileName: "sample.tf",
	}
	sample := `provider "google" {
		version = "2.5.0"
		project = "instrumenta"
		region = "europe-west2"
	  
	  
	  }
	  
	  resource "google_container_cluster" "primary" {
		name     = "my-gke-cluster"
		location = "us-central1"
	  
		# We can't create a cluster with no node pool defined, but we want to only use
		# separately managed node pools. So we create the smallest possible default
		# node pool and immediately delete it.
		remove_default_node_pool = true
		initial_node_count = 1
	  
		# Setting an empty username and password explicitly disables basic auth
		master_auth {
		  username = ""
		  password = ""
		}
	  }
	  
	  resource "google_container_node_pool" "primary_preemptible_nodes" {
		name       = "my-node-pool"
		location   = "us-central1"
		cluster    = "${google_container_cluster.primary.name}"
		node_count = 1
	  
		node_config {
		  preemptible  = true
		  machine_type = "n1-standard-1"
	  
		  metadata = {
			disable-legacy-endpoints = "true"
		  }
	  
		  oauth_scopes = [
			"https://www.googleapis.com/auth/logging.write",
			"https://www.googleapis.com/auth/monitoring",
		  ]
		}
	  }
	  
	  # The following outputs allow authentication and connectivity to the GKE Cluster
	  # by using certificate-based authentication.
	  output "client_certificate" {
		value = "${google_container_cluster.primary.master_auth.0.client_certificate}"
	  }
	  
	  output "client_key" {
		value = "${google_container_cluster.primary.master_auth.0.client_key}"
	  }
	  
	  output "cluster_ca_certificate" {
		value = "${google_container_cluster.primary.master_auth.0.cluster_ca_certificate}"
	  }
	  `

	var input interface{}
	err := parser.Unmarshal([]byte(sample), &input)
	if err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("There should be information parsed but its nil")
	}

	inputMap := input.(map[string]interface{})
	if len(inputMap["Resources"].([]interface{})) <= 0 {
		t.Error("There should be resources defined in the parsed file, but none found")
	}
}
