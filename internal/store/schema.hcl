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
  column "write_token" {
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

table "folders" {
  schema = schema.main
  column "id" {
    null = false
    type = text
  }
  primary_key {
    columns = [column.id]
  }
}

table "feeds" {
  schema = schema.main
  column "id" {
    null = false
    type = text
  }
  column "title" {
    null = false
    type = text
  }
  column "url" {
    null = false
    type = text
  }
  column "htmlUrl" {
    null = false
    type = text
  }
  primary_key {
    columns = [column.id]
  }
}

table "feed_folders" {
  schema = schema.main
  column "feed_id" {
    null = false
    type = text
  }
  column "folder_id" {
    null = false
    type = text
  }
  primary_key {
    columns = [column.feed_id, column.folder_id]
  }
  foreign_key "0" {
    columns     = [column.folder_id]
    ref_columns = [table.folders.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  foreign_key "1" {
    columns     = [column.feed_id]
    ref_columns = [table.feeds.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  index "idx_feed_folders_by_folder" {
    columns = [column.folder_id]
  }
  index "idx_feed_folders_by_feed" {
    columns = [column.feed_id]
  }
}

table "articles" {
  schema = schema.main
  column "id" { // timestamp_usec
    null = false
    type = text
  }
  column "feed_id" {
    null = false
    type = text
  }
  column "title" {
    null = false
    type = text
  }
  column "content" {
    null = true
    type = text
  }
  column "author" {
    null = true
    type = text
  }
  column "href" {
    null = true
    type = text
  }
  column "published_at" {
    null = true
    type = int
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "0" {
    columns     = [column.feed_id]
    ref_columns = [table.feeds.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  index "idx_articles_feed_id" {
    columns = [column.feed_id]
  }
  index "idx_articles_published" {
    on {
      desc   = true
      column = column.published_at
    }
  }
  index "idx_articles_feed_published" {
    on {
      column = column.feed_id
    }
    on {
      desc   = true
      column = column.published_at
    }
  }
}

table "article_statuses" {
  schema = schema.main
  column "article_id" {
    null = false
    type = text
  }
  column "is_read" {
    null    = false
    type    = boolean
    default = 0
  }
  column "is_starred" {
    null    = false
    type    = boolean
    default = 0
  }
  primary_key {
    columns = [column.article_id]
  }
  foreign_key "0" {
    columns     = [column.article_id]
    ref_columns = [table.articles.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  index "idx_article_statuses_read" {
    columns = [column.is_read]
  }
  index "idx_article_statuses_starred" {
    columns = [column.is_starred]
  }
}

table "pending_actions" {
  schema = schema.main
  column "id" {
    null           = false
    type           = integer
    auto_increment = true
  }
  column "article_id" {
    null = false
    type = text
  }
  column "action" {
    null = false
    type = text
  }
  column "created_at" {
    null    = false
    type    = integer
    default = sql("strftime('%s', 'now')")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "0" {
    columns     = [column.article_id]
    ref_columns = [table.articles.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  index "idx_pending_actions_created_at" {
    columns = [column.created_at]
  }
  index "idx_pending_actions_article_id" {
    columns = [column.article_id]
  }
  check {
    expr = "(action IN ('read', 'unread', 'star', 'unstar'))"
  }
}
