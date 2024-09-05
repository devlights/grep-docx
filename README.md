# これは何？

指定ディレクトリの下に存在する MS WORD ファイル (*.docx) をGREPコマンドのように検索します。

対象のWORDファイルが存在していて、どこに書いてあるのかが分からない時などに良かったらご利用ください。

Windowsのみで動作します。

## インストール

```sh
go install github.com/grep-docx/cmd/grep-docx@latest
```

## 使い方

```sh
$ ./grep-docx.exe -h
Usage of ./grep-docx.exe:
  -debug
        debug mode
  -dir string
        directory (default ".")
  -json
        output as JSON
  -only-hit
        show ONLY HIT (default true)
  -text string
        search text
  -verbose
        verbose mode
```

ヒットした文書のパスが知りたい場合は以下のようにします。

```sh
$ ./grep-docx.exe -dir ~/path/to/documents -text "データベースサイズ"
test.docx: HIT
```

ヒットした箇所も見たい場合は ```-verbose``` オプションを付与するとみることが出来ます。

```sh
$ ./grep-docx.exe -dir ~/path/to/documents -text "データベース*サイズ" -verbose  
```

結果をjsonで出力したい場合は ```-json``` オプションを付与します。

```sh
$ ./grep-docx.exe -dir ~/path/to/documents -text "データベース*サイズ" -verbose  -json
```

## ビルド方法

[Task](https://taskfile.dev/#/) を使っています。詳細は [Taskfile.yml](./Taskfile.yml) を参照ください。

```sh
$ task build
```
