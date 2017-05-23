job "dp-search-query" {
  datacenters = ["DATA_CENTER"]
  constraint {
  }
  update {
    stagger = "10s"
    max_parallel = 1
  }
  group "dp-search-query" {
    task "dp-search-query" {
      artifact {
        source = "s3::S3_TAR_FILE"
        destination = "."
        // The Following options are needed if no IAM roles are provided
        // options {
        // aws_access_key_id = ""
        // aws_access_key_secret = ""
        // }
      }
      env {
        ELASTIC_URL = "ELASTIC_SEARCH_URL"
        PORT = "$NOMAD_PORT_http"
      }
      driver = "exec" // To run on OSX change this to raw_exec
      config {
        command = "bin/dp-search-query"
        args = []
      }
      resources {
        cpu = 600
        memory = 400
        network {
          port "http" {}
        }
        }
      service {
          port = "HEALTHCHECK_PORT"
          check {
              type     = "http"
              path     = "HEALTHCHECK_ENDPOINT"
              interval = "10s"
              timeout  = "2s"
          }
        }
      }
    }
}
