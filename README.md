# go_homepage
A static website generator for https://stevenchung.ca written with Go and Webpack.

It has:
- A homepage of blog post listings
- Blog posts written in Markdown
- An atom feed of blog posts
- Reading page full of Goodreads reviews
- About page written in Markdown

Goodreads reviews are retrieved via API and cached locally. All JS, CSS, and images are compressed via Webpack. Images are also responsive.

## Requirements
- [Docker Desktop](https://www.docker.com) (if no Docker, see [`Dockerfile`](Dockerfile))
- [direnv](https://github.com/direnv/direnv) to automatically load/unload environment variables when inside of this project or you can [load them yourself](.example.envrc)

## Run Locally
Setup docker via:
```sh
make docker-install
```

This will build docker and do all the installation inside docker. After, it’ll copy all downloaded code libraries from the docker container to the host, so that when the container and host sync filesystems, the libraries will be there.

Look for all files with `.example` in the filename. Make a copy without `.example` and fill in the missing information. You can skip the [`aws`](aws) folder, as that relates to deploying. You also can set more settings by looking at [`settings.go`](settings/settings.go).

You can run a developer instance through Docker via:
```sh
make docker
```

For production:
```sh
make docker-prod
```

By default, you can access the file server via `http://localhost:3000`.

All blog post content is in the `content` folder.

## Code Structure
`go_homepage` manages the following through a `Makefile`:

- A Go executable
- A Webpack setup
- `watchman` - a file diff watcher by Facebook to auto compile everything conveniently for development
- `aws-cli` - to handle Amazon S3 Deployment

First, Webpack compiles/optimizes all the assets (JS, CSS, images) in the `assets` and `content` folders (`make build-assets`). It also generates a `Manifest.json` and `content/responsive` folder of json files. These files give the Go executable file paths from Webpack compilation.

After, the Go executable generates all the web page files using the content folder and the json files from Webpack (`make build-go`). It also does any API requests if needed. By default, the generated results are put in the `generated` folder. The Go executable can also host these files with a file server (`make file-server`).

The code is structured much like a Go web app. In fact, you can start it as a web app (`make server`), which uses Go standard lib `net/http` internally.

I run them through Docker, which handles all the system dependencies to install Go, nodejs, image optimization, etc. See [`Dockerfile`](Dockerfile) to see system dependencies. Your local system probably has some of them already, as Docker is running Alpine, a minimal Linux distribution.

## Deploy
`go_homepage` is designed to be hosted on Amazon S3.

### Setup
Start by creating an Amazon S3 bucket with a bucket name, we’ll call `<S3_BUCKET>`. I recommend calling the bucket your domain name (`www.yourwebsite.com` or `yourwebsite.com`). Then go to `Amazon IAM > Users`. Add a new User and give him `Programmatic Access` and give him the following policy:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "aws-cli policy",
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetObject",
                "s3:ListBucket",
                "s3:DeleteObject"
            ],
            "Resource": [
                "arn:aws:s3:::<S3_BUCKET>/*",
                "arn:aws:s3:::<S3_BUCKET>"
            ]
        }
    ]
}
```

Store the `Access Key ID` and `Secret Access Key` in `aws/credentials` of the project (you can’t retrieve the `Secret Access Key` later). `aws-cli` will use this.

### Deploy
Make sure you’re able to run everything locally and you can upload all your files to S3 via:

```sh
make docker-deploy
```

I usually use:

```sh
make push-docker-deploy
```

because it ensures my origin master is synced with my homepage. In the future, I’ll make `go_homepage` deploy via Travis CI, so whenever origin master updates, the homepage updates.

## Host
You can host `go_homepage` directly from Amazon S3, it’s easy to Google for.

I find it best to use Amazon CloudFront CDN with Amazon Certificate Manager to provide SSL.

### Setup
First, open the Amazon Certificate Manager and request a public certificate. Use both `*.yourwebsite.com` and `yourwebsite.com` as domain names. Use DNS validation, which will tell you to create a `CNAME` record in your DNS for your domain with the given `name` and `value` (note that `name` is the full url host, your DNS may only need the subdomain of the `name` to specify the host).

Create that `CNAME` record and wait until Amazon Certificate Manager validates.

Once validated, go to Amazon CloudFront and create a distribution. Choose `Web` and:

- Link CloudFront and S3 privately through an Origin Access Identity:
    - Origin Domain Name: `<S3_BUCKET>`
    - Restrict Bucket Access: Yes
    - Origin Access Identity: Create a new identity
    - Grant Read Permissions on Bucket: Yes, Update Bucket Policy (this will change the bucket policy on S3 for CloudFront access)
- Allow access through another domain:
    - Alternate Domain Names (CNAMEs): yourwebsite.com (url you’re hosting)
- Default Root Object: index.html
- SSL Certificate: Custom SSL Certificate, choose the Amazon Certificate Manager certificate

These are the **must have** options, I’d browse through the other options and see what you like.

After creation and when status is deployed (a few minutes), you can access your website via the `Domain Name` (<some_hash>.cloudfront.net). Create a `CNAME` record on your DNS to point the url you’re hosting (yourwebsite.com) to that `Domain Name`.