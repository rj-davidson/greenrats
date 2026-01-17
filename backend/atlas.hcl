variable "database_url" {
  type    = string
  default = "postgres://greenrats:greenrats@localhost:5432/greenrats?sslmode=disable"
}

env "local" {
  src = "ent://ent/schema"
  dev = "docker://postgres/16/dev?search_path=public"
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
  url = var.database_url
}

env "production" {
  src = "ent://ent/schema"
  migration {
    dir = "file://migrations"
  }
  url = var.database_url
}
