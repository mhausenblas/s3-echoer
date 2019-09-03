# S3 echoer

A simple demo tool that reads input from `stdin` and uploads it to an
exiting S3 bucket, keyed by the creation timestamp.

- [Install it](#install-it)
- [Use it](#use-it)
  - [Prepare S3](#prepare-s3)
  - [Locally](#locally)
  - [Kubernetes](#kubernetes)
  - [Verify S3 write](#verify-s3-write)

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

### Kubernetes

Running `s3-echoer` in-cluster means that we somehow need to give it the 
permission to write to S3. There are two approaches to this: granting the node 
the pod runs on the permission or granting the pod the permission, directly.

#### Node-level approach

With the node-level approach, we grant all nodes the permission to write to S3.
For this, we need to look up the respective role and attach the `AmazonS3FullAccess` 
policy like shown in the following.

The EKS cluster named `s3echoer` was provisioned with `eksctl` and hence we're looking
for a node group called `eksctl-s3echoer-nodegroup-ng`:

```sh
$ INSTANCE_PROFILE_PREFIX=$(aws cloudformation describe-stacks | jq -r '.Stacks[].StackName' | grep eksctl-s3echoer-nodegroup-ng)
$ INSTANCE_PROFILE_NAME=$(aws iam list-instance-profiles | jq -r '.InstanceProfiles[].InstanceProfileName' | grep $INSTANCE_PROFILE_PREFIX)
$ ROLE_NAME=$(aws iam get-instance-profile --instance-profile-name $INSTANCE_PROFILE_NAME | jq -r '.InstanceProfile.Roles[] | .RoleName')
$ aws iam attach-role-policy \
      --role-name $ROLE_NAME \
      --policy arn:aws:iam::aws:policy/AmazonS3FullAccess
```

Next, create the service account `s3-echoer` that the app will be using:

```sh
$ kubectl create sa s3-echoer
```

Now you can launch the job like so:

```sh
$ sed -e "s/TARGET_BUCKET/${TARGET_BUCKET}/g" s3-echoer-job.yaml.template > s3-echoer-job.yaml
$ kubectl apply -f s3-echoer-job.yaml
```

You can check how the job did using:

```sh
$ kubectl logs job/s3-echoer
Uploading user input to S3 using s3-echoer-demo/s3echoer-1565024447
```

Obviously, this approach has a critical security drawback: it grants every pod 
running on any of the nodes the right to write to S3, violating the least-privileges 
principle.

We can do better with the pod-level approach.

#### Pod-level approach

Using IAM roles for service accounts (IRSA) as per [#23](https://github.com/aws/containers-roadmap/issues/23) 
we can grant individual apps certain permissions. Let's do that for our case.

Again, assuming the EKS cluster named `s3echoer` was provisioned with `eksctl`,
we first enable IRSA in the cluster like so:

```sh
$ eksctl utils associate-iam-oidc-provider --name s3echoer --approve
```

And now we can create the service account `s3-echoer` used by the app and along
with it create a role that has the `AmazonS3FullAccess` policy attached:

```sh
$ eksctl create iamserviceaccount \
                --name s3-echoer \
                --cluster s3echoer \
                --attach-policy-arn arn:aws:iam::aws:policy/AmazonS3FullAccess \ 
                --approve
```

Now we have the permissions in place, the pod needs to acquire the temporary
credentials via [STS](https://docs.aws.amazon.com/STS/latest/APIReference/Welcome.html).
This can be achieved either via using an [IRSA-enabled SDK](https://github.com/aws/aws-sdk-go/releases/tag/v1.21.9) 
as [is the case here](https://github.com/mhausenblas/s3-echoer/blob/a8309a704a40ad67ad5f0ac3685e198c97e7bd6b/main.go#L80),
or manually, by calling `aws sts assume-role-with-web-identity` like so:

```sh
$ JQ=/usr/bin/jq && curl https://stedolan.github.io/jq/download/linux64/jq > $JQ && chmod +x $JQ
$ curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py && python get-pip.py && pip install awscli --upgrade
$ aws sts assume-role-with-web-identity \
          --role-arn $AWS_ROLE_ARN \
          --role-session-name mh9test \
          --web-identity-token file://$AWS_WEB_IDENTITY_TOKEN_FILE \
          --duration-seconds 1000 > /tmp/irp-cred.txt
$ export AWS_ACCESS_KEY_ID="$(cat /tmp/irp-cred.txt | jq -r ".Credentials.AccessKeyId")"
$ export AWS_SECRET_ACCESS_KEY="$(cat /tmp/irp-cred.txt | jq -r ".Credentials.SecretAccessKey")"
$ export AWS_SESSION_TOKEN="$(cat /tmp/irp-cred.txt | jq -r ".Credentials.SessionToken")"
$ rm /tmp/irp-cred.txt
```

Now you can launch the job like so:

```sh
$ sed -e "s/TARGET_BUCKET/${TARGET_BUCKET}/g" s3-echoer-job.yaml.template > s3-echoer-job.yaml
$ kubectl apply -f s3-echoer-job.yaml
```

You can check how the job did using:

```sh
$ kubectl logs job/s3-echoer
Uploading user input to S3 using s3-echoer-demo/s3echoer-1565024447
```

Note: you can clean up with `kubectl delete job/s3-echoer`.

### Verify S3 write

Let's check if the data landed in the right place:

```sh
$ aws s3api list-objects \
            --bucket $TARGET_BUCKET \
            --query 'Contents[].{Key: Key, Size: Size}'
[
    {
        "Key": "s3echoer-1563907403",
        "Size": 63
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
This is an in-cluster test
```

And that's it :)



