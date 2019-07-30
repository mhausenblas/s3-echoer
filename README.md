# S3 echoer

A simple demo tool that reads input from `stdin` and uploads it to an
exiting S3 bucket, keyed by the creation timestamp.

- [Install it](#install-it)
- [Use it](#use-it)
  - [Prepare S3](#prepare-s3)
  - [Locally](#locally)
  - [Kubernetes](#kubernetes) 

## Install it

To install `s3-echoer`, execute the following two commands. Download the 
respective binary for your platform from the [releases page](https://github.com/mhausenblas/s3-echoer/releases), here shown for `macOS`, and make it executable:

```sh
$ curl -L https://github.com/mhausenblas/s3-echoer/releases/latest/download/s3-echoer-macos -o /usr/local/bin/s3-echoer

$ chmod +x /usr/local/bin/s3-echoer
```

## Use it

### Prepare S3

Make sure the bucket you want to write to exists, for example, let's assume we 
want to write to a bucket called `s3-echoer-demo`. So first we define the target 
bucket using an environment variable like so:

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

### Locally

Now that we've made sure the S3 bucket exists, let's use it:

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

### Kubernetes

Create a service account:

```sh
$ kubectl create sa s3-echoer
```

Now you can launch the job like so:

```sh
$ kubectl apply -f s3-echoer-job.yaml
```

Finally, we jump into the pod and execute `s3-echoer` directly in there:

```sh
$ JUMPOD=$(kubectl get po -l=job-name=s3-echoer --output=jsonpath={.items[*].metadata.name})
$ kubectl exec -it $JUMPOD sh
```

Once in the container, do the following:

```sh
sh-4.2# echo This is an in-cluster test | /s3-echoer s3-echoer-demo
Uploading user input to S3 using s3-echoer-demo/s3echoer-1563972036
```

Note, above, that in order to trigger the upload to S3 you need to press `ENTER` and then `CTRL+D`.

From here on you can again use `aws s3api list-objects` and/or `aws s3api get-object` to verify if the write to S3 worked, or, alternatively, use the
AWS console for this.

Node-level approach (incl. implications):

- determine role of nodes `ROLE_NAME`
- `aws iam put-role-policy` with write permissions to bucket

Pod-level approach:

TBD and also compare and discuss approaches concerning least privileges and attack surface.