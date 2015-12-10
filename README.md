Key Image Store
===============

Self hosting image store service.

Get started
-----------
```
go get
go run kis.go
```

API specification
-----------------
### Save image
#### Request
* endpoint    : `/api/`
* method      : POST
* Content-Type: `multipart/form-data` or other
* keyname     : `image`
* filetype    : GIF or PNG or JPEG or WEBP

※ If Content-Type isn't `multipart/form-data` then server get image file from request body and judge a MIME type of file.

#### Response
* Content-Type: `application/json`
* Example (200)
````
{"status": "ok"}
````
* Example (400)
```
{
  "status": "ng",
  "error": "bad request"
}
```

```sh
curl "http://localhost:3000/api/" --verbose -F "image=@test.jpg"
# or
curl "http://localhost:3000/api/" --verbose -H "Content-Type: image/jpeg" --data-binary "@test.jpg"
```

### Fetch image
#### Request
* endpoint      : `/api/:id.ext`
* method        : GET
* extension     : one of `gif, png, jpg, webp, json`

#### Response
* Content-Type: `image/gif` or `image/png` or `image/jpeg` or `image/webp` or `application/json`
* Example (200) metadata `/api/:id.json`
```
{
  "status": "ok",
  "width": 600,
  "height": 400
}
```
* Example (404)
```
{
  "status": "ng",
  "error": "file not found"
}
```

```sh
curl "http://localhost:3000/api/963a1da7-08bd-48ca-9e1a-53f98ba06e39.json" --verbose
```
