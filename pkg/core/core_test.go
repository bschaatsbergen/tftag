package core

import (
	"fmt"
	"testing"

	"github.com/bschaatsbergen/tftag/pkg/model"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func Test_processHCLFile(t *testing.T) {
	type args struct {
		config   model.Config
		fileName string
		body     string
	}

	tests := []struct {
		name   string
		args   args
		expect string
	}{
		{
			name: "add tags to resource without tags",
			args: args{
				config: model.Config{Config: []model.TfTagConfig{
					{"all", map[string]string{"Pine": "Apple"}},
				}},
				fileName: "",
				body: `
resource "aws_s3_bucket" "users" {
  bucket = "users-bucket"
}

resource "google_workstations_workstation_cluster" "test" {
  workstation_cluster_id = "workstation-cluster"
}
`,
			},
			expect: `
resource "aws_s3_bucket" "users" {
  bucket = "users-bucket"
  tags = {
    Pine = "Apple"
  }
}

resource "google_workstations_workstation_cluster" "test" {
  workstation_cluster_id = "workstation-cluster"
  labels = {
    Pine = "Apple"
  }
}
`,
		},
		{
			name: "add tags to resource with tags",
			args: args{
				config: model.Config{Config: []model.TfTagConfig{
					{"all", map[string]string{"Pine": "Apple"}},
				}},
				fileName: "",
				body: `
resource "aws_s3_bucket" "users" {
  bucket = "users-bucket"
  tags = {
    BusinessUnit = "Finance"
  }
}

resource "google_workstations_workstation_cluster" "test" {
  workstation_cluster_id = "workstation-cluster"
  labels = {
    BusinessUnit = "Finance"
  }
}
`,
			},
			expect: `
resource "aws_s3_bucket" "users" {
  bucket = "users-bucket"
  tags = {
    BusinessUnit = "Finance"
    Pine         = "Apple"
  }
}

resource "google_workstations_workstation_cluster" "test" {
  workstation_cluster_id = "workstation-cluster"
  labels = {
    BusinessUnit = "Finance"
	Pine         = "Apple"
  }
}
`,
		},
		{
			name: "override tags on resource with tags",
			args: args{
				config: model.Config{Config: []model.TfTagConfig{
					{"all", map[string]string{"Pine": "Apple"}},
				}},
				fileName: "",
				body: `
resource "aws_s3_bucket" "users" {
  bucket = "users-bucket"
  tags = {
    Pine         = "Tree"
    BusinessUnit = "Finance"
  }
}

resource "google_workstations_workstation_cluster" "test" {
  workstation_cluster_id = "workstation-cluster"
  labels = {
	Pine         = "Tree"
    BusinessUnit = "Finance"
  }
}
`,
			},
			expect: `
resource "aws_s3_bucket" "users" {
  bucket = "users-bucket"
  tags = {
    BusinessUnit = "Finance"
    Pine         = "Apple"
  }
}

resource "google_workstations_workstation_cluster" "test" {
  workstation_cluster_id = "workstation-cluster"
  labels = {
    Pine         = "Tree"
    BusinessUnit = "Finance"
  }
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Printf("processing %s", tt.name)
			file, diagnostics := hclwrite.ParseConfig([]byte(tt.args.body), tt.args.fileName, hcl.Pos{Line: 1, Column: 1})
			if diagnostics != nil {
				t.Fatalf("%v", diagnostics)
			}
			processHCLFile(file, tt.args.config, tt.args.fileName)
			result := string(hclwrite.Format(file.Bytes()))
			if result != tt.expect {
				t.Errorf("expected: %s, got: %s", tt.expect, result)
			}
		})
	}
}
