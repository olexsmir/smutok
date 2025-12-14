schema "main" {}

table "reader" {
  schema = schema.main
  column "id" {
    null           = true
    type           = integer
    auto_increment = true
  }
  column "token" {
    null = true
    type = text
  }
  column "last_sync" {
    null = true
    type = date
  }
  primary_key {
    columns = [column.id]
  }
}
