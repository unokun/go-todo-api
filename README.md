プログラミング言語goを使ってWeb API開発する記事はいろいろとありますが、できるだけ外部パッケージを使わない方針で開発しましたので報告します。goを使ったWeb API開発する際の参考になれば幸いです。
また、Web APIを公開する場合、Webサーバー(nginx) + アプリケーションサーバー(go)という構成が一般的(?)かと思いましたので、設定についての記載を最後に追加しました。

ソースコードはGithubにあります。
https://github.com/unokun/go-todo-api

## 環境
| 項目       | バージョン |
| ---------- | ------ |
| OS         | Centos7|
| go              | 1.11.4 |
| mysql(maria DB) | 5.5.60 |

## TODO管理 Web API
baseurl(/todo/api/v1)にPathを加えたURLで各機能を実装しています。

| Path       | HTTPメソッド |機能|
| ---------- | ------ | -------------- |
| /tasks     | GET    | タスク一覧取得 |
| /tasks     | POST   | タスク登録     |
| /tasks/:id | PATCH  | タスク更新     |
| /tasks/:id | DELETE | タスク削除     |

### インストール
goのインストールは、以下の記事を、
[CentOSにGo言語のインストール \- Qiita](https://qiita.com/estaro/items/23c10fe7e43a2a70e689)

また、Web APIの実装については以下の記事を参考にさせていただきました。
[GoでWebアプリを作ろう 第一回 : Goで簡単なCRUD \- Studio Andy](http://studio-andy.hatenablog.com/entry/go-todo-crud)

使った外部パッケージは、以下の二つです。
* gin
* mysql

[gin](https://github.com/gin-gonic/gin)は、go製のWebフレームワークです。jsonレスポンス処理機能も持っている優れものです。
mysqlはDBに合わせて変更してください。

```shell
$ go get github.com/gin-gonic/gin
$ go get github.com/go-sql-driver/mysql
```

### ファイル構成
```shell
$ tree go-todo-api
go-todo-api
├── db
│   └── createTask.sql
└── src
    ├── controller
    │   └── task.go
    ├── main.go
    └── model
        ├── model.go
        └── task.go
```

### DB
データベース(tododb)作成
専用のユーザー(todo)を作成し権限を付与します。

```shell
$ mysql -u root -p
MariaDB [(none)]> create database tododb default charset utf8;
MariaDB [(none)]> create user 'todo'@'localhost' identified by 'todo';
MariaDB [(none)]> grant all on `tododb`.* to 'todo'@'localhost';
```
テーブルはタイトルのみ持つシンプルな構成です。

```sql
USE tododb
DROP TABLE IF EXISTS task;
CREATE TABLE IF NOT EXISTS task (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title VARCHAR(255) NOT NULL,
    PRIMARY KEY(id)
);
```

### model
DBアクセス処理を作成します。
task.goにJson返却用のstructを定義しています。

```golang
package model

import (
    "time"
)

type Task struct {
    ID        uint      `json:"id"`         // id
    CreatedAt time.Time `json:"created_at"` // created_at
    UpdatedAt time.Time `json:"updated_at"` // updated_at
    Title     string    `json:"title"`      // title
}
```
model.goにDB接続用関数を作成しています。

```golang
package model

import (
    "database/sql"
    "log"
    "os"

    // mysql driver
    _ "github.com/go-sql-driver/mysql"
)

// DBConnect returns *sql.DB
func DBConnect() (db *sql.DB) {
    dbDriver := "mysql"
    dbUser := "todo"
    dbPass := os.Getenv("MYSQL_TODO_PASSWORD") // 環境変数から取得
    dbName := "tododb"
    dbOption := "?parseTime=true"
    db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName+dbOption)
    if err != nil {
        log.Fatal(err)
    }
    return db
}
```
### controller
mainから呼び出すタスク処理を記述します。

```golang
package controller

import (
    "fmt"
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/unokun/go-todo-api/src/model"
)

// タスク一覧
func TasksGET(c *gin.Context) {
    db := model.DBConnect()
    result, err := db.Query("SELECT * FROM task ORDER BY id DESC")
    if err != nil {
        panic(err.Error())
    }

    // json返却用
    tasks := []model.Task{}
    for result.Next() {
        task := model.Task{}
        var id uint
        var createdAt, updatedAt time.Time
        var title string

        err = result.Scan(&id, &createdAt, &updatedAt, &title)
        if err != nil {
            panic(err.Error())
       }

        task.ID = id
        task.CreatedAt = createdAt
        task.UpdatedAt = updatedAt
        task.Title = title
        tasks = append(tasks, task)
    }
    c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}
// タスク検索
func FindByID(id uint) model.Task {
    db := model.DBConnect()
    result, err := db.Query("SELECT * FROM task WHERE id = ?", id)
    if err != nil {
        panic(err.Error())
    }
    task := model.Task{}
    for result.Next() {
        var createdAt, updatedAt time.Time
        var title string

        err = result.Scan(&id, &createdAt, &updatedAt, &title)
        if err != nil {
            panic(err.Error())
        }

        task.ID = id
        task.CreatedAt = createdAt
        task.UpdatedAt = updatedAt
        task.Title = title
    }
    return task
}
// タスク登録
func TaskPOST(c *gin.Context) {
    db := model.DBConnect()

    title := c.PostForm("title")
    now := time.Now()

    _, err := db.Exec("INSERT INTO task (title, created_at, updated_at) VALUES(?, ?, ?)", title, now, now)
    if err != nil {
        panic(err.Error())
    }

    fmt.Printf("post sent. title: %s", title)
}
// タスク更新
func TaskPATCH(c *gin.Context) {
    db := model.DBConnect()

    id, _ := strconv.Atoi(c.Param("id"))
    title := c.PostForm("title")
    now := time.Now()

    _, err := db.Exec("UPDATE task SET title = ?, updated_at=? WHERE id = ?", title, now, id)
    if err != nil {
        panic(err.Error())
    }

    task := FindByID(uint(id))

    fmt.Println(task)
    c.JSON(http.StatusOK, gin.H{"task": task})
}
// タスク削除
func TaskDELETE(c *gin.Context) {
    db := model.DBConnect()

    id, _ := strconv.Atoi(c.Param("id"))

    _, err := db.Query("DELETE FROM task WHERE id = ?", id)
    if err != nil {
        panic(err.Error())
    }

    c.JSON(http.StatusOK, "deleted")
}
```
### main
メイン処理を記述します。

```golang
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/unokun/go-todo-api/src/controller"
)

func main() {
    router := gin.Default()

    v1 := router.Group("/todo/api/v1")
    {
        v1.GET("/tasks", controller.TasksGET)
        v1.POST("/tasks", controller.TaskPOST)
        v1.PATCH("/tasks/:id", controller.TaskPATCH)
        v1.DELETE("/tasks/:id", controller.TaskDELETE)
    }
    // nginxのreverse proxy設定
    router.Run(":9000")
}
```

### 実行
バックグランドジョブとして実行します。

```shell
$ MYSQL_TODO_PASSWORD=xxx go run main.go &
```

### API呼び出し
API呼び出しを実行する方法はいろいろありますが、Google Chromeの拡張機能である「Restlet Client」が使いやすいです。
[Restlet Client \- REST API Testing \- Chrome ウェブストア](https://chrome.google.com/webstore/detail/restlet-client-rest-api-t/aejoelaoggembcahagimdiliamlcdmfm?hl=ja)

タスク一覧
![go_todo_list_tasks.png](https://qiita-image-store.s3.amazonaws.com/0/5247/ad886d79-963b-a595-8c07-83210b75ba11.png)


タスク登録
![go_todo_add_task.png](https://qiita-image-store.s3.amazonaws.com/0/5247/6358e309-3236-d4b1-8f0d-05c15fc13663.png)


ginのログ
![gin_log.png](https://qiita-image-store.s3.amazonaws.com/0/5247/105a4060-24dc-1cbc-3937-b42034981d17.png)

## nginxの設定
Webサーバー(nginx) + アプリケーションサーバー(go webapi)という構成にする場合、リバースプロキシーとして動作させます。locationの記述については[ここ](https://server-setting.info/centos/nginx-location-check.html)を参考にしてください。

```
$ cat /etc/nginx/conf.d/https.conf
        location ^~ /todo/ {
                proxy_pass   http://127.0.0.1:9000;
        }
```

## リンク
* [CentOSにGo言語のインストール \- Qiita](https://qiita.com/estaro/items/23c10fe7e43a2a70e689)
* [GoでWebアプリを作ろう 第一回 : Goで簡単なCRUD \- Studio Andy](http://studio-andy.hatenablog.com/entry/go-todo-crud)
* [nginxでgoを動かす \- 冷やしブログはじめました](http://pokrkami.hatenablog.com/entry/2016/01/27/233048)
* [Nginx で location の判定方法と優先順位を調べる \| レンタルサーバー・自宅サーバー設定・構築のヒント](https://server-setting.info/centos/nginx-location-check.html)
