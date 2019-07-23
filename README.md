# S3 echoer

A simple demo tool that reads input from `stdin` and uploads it to an
exiting S3 bucket, keyed by the creation timestamp.

Make sure the bucket you want to write to exists, for example, say we want
to write to a bucket called `s3-echoer-demo`.

Let's first define the target bucket using an environment variable:

```sh
$ TARGET_BUCKET=s3-echoer-demo
```

Now you can check if the S3 bucket exists:

```sh
$ aws s3 ls | grep $TARGET_BUCKET
```

If the S3 bucket doesn't exist yet, create it like so:

```sh
$ aws s3api create-bucket \
            --bucket $TARGET_BUCKET \
            --create-bucket-configuration LocationConstraint=$(aws configure get region) \
            --region $(aws configure get region)
```

OK, now that we've made sure the S3 bucket exists, let's use it. First you need
to download the binary and put it somewhere on your path. And then do:

```sh
$ s3-echoer $TARGET_BUCKET
This is a test. 
And it should land in the target bucket ...
CTRL+D
Uploading user input to S3 using s3-echoer-demo/s3echoer-1563906471
```

Now let's check if it landed in the right place:

```sh
$ aws s3api list-objects \
            --bucket $TARGET_BUCKET \
            --query 'Contents[].{Key: Key, Size: Size}'
[
    {
        "Key": "s3echoer-1563907403",
        "Size": 60
    }
]
```

Yay, that worked like a charm! Seems an S3 object with our content has been 
created, in the target bucket. And as a final check, let's download the object
and store it in a file to check if it actually contains the text we entered, above:

```sh
$ aws s3api get-object \
            --bucket $TARGET_BUCKET \
            --key s3echoer-1563907403 \
            /tmp/s3echoer-readback.txt

$ cat /tmp/s3echoer-readback.txt
This is a test.
And it should land in the target bucket ...

```

And that's it :)