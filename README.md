# File Upload Server

This is a simple Go application that provides an HTTP server to handle file uploads. The server can accept file uploads and store them to an object storage service called Minio. The server provides a root endpoint to check if the server is online and a file upload endpoint.
## Getting Started

These instructions will help you run the application.
## Prerequisites


- Go 1.16 or later
- Minio server


## Installation

1. Clone the repository:

```bash
git clone https://github.com/URFU-2022-machine-learning-engineering/speech-recognition-API
```
2. Set the following environment variables:

- MINIO_ENDPOINT: the Minio endpoint URL
- MINIO_ACCESS_KEY: the Minio access key
- MINIO_SECRET_KEY: the Minio secret key
- MINIO_BUCKET: the Minio bucket to upload files to

3. Build the application:

```bash
go build
```

4. Run the application:

```bash
./file-upload-server
```
## Usage

The server provides two endpoints:

## Root Endpoint

Returns a message indicating that the server is online.

```
GET /
```

## File Upload Endpoint

Accepts file uploads and stores them to the Minio bucket specified in the environment variables.

```bash
POST /upload
```

The file should be uploaded as a multipart form data with the field name file. If the upload is successful, the server will return a JSON response with a message indicating success. If the upload fails, the server will return a JSON response with a message indicating the failure reason.

## License

This project is licensed under the MIT License - see the LICENSE file for details.