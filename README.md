# Kubernetes CloudFormation Signal

[![Build Status](https://travis-ci.org/UKHomeOffice/kube-cfn-signal.svg?branch=master)](https://travis-ci.org/UKHomeOffice/kube-cfn-signal)

This little utility can health check kubernetes endpoints until they become
ready and send a signal to CloudFormation API.

CloudFormation allows you to set
[CreationPolicy](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-attribute-creationpolicy.html)
and
[UpdatePolicy](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-attribute-updatepolicy.html)
attributes on stack resources, the one we're interested in is the Autoscaling group
resource which looks after the Kubernetes nodes.

The most useful place to use this is when you're doing AutoScaling group
rolling updates.

### Requirements
#### IAM Instance Policy
Normally you would want to run `kube-cfn-signal` from within an instance which
is being created/updated. So to make things simpler, it is advisable to allow
your kubernetes nodes to query tags and send a signal to CloudFormation API.

```json
{
    "Statement": [
        {
            "Resource": "arn:aws:ec2:*:*:instance/*",
            "Action": [
                "ec2:DescribeTags",
            ],
            "Effect": "Allow"
        },
        {
            "Resource": "arn:aws:cloudformation:*:*:stack/*/*",
            "Action": [
                "cloudformation:SignalResource"
            ],
            "Effect": "Allow"
        }
    ]
}
```

### Running
#### Systemd Unit
```
[Unit]
Description=Kubernetes cfn signal
Documentation=https://github.com/UKHomeOffice/kube-cfn-signal

[Service]
Type=oneshot
PrivateTmp=true
ProtectSystem=full
RemainAfterExit=yes
TimeoutStartSec=10m
ExecStart=/opt/bin/kube-cfn-signal --insecure-skip-tls-verify
```

## Build

Dependencies are located in the vendor directory and managed using
[govendor](https://github.com/kardianos/govendor) cli tool.

```
go test -v -cover

mkdir -p bin
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.Version=dev+git" -o bin/kube-cfn-signal
```


## Release process

Push / Merge to master will produce a docker
[image](https://quay.io/repository/ukhomeofficedigital/kube-cfn-signal?tab=tags) with a tag `latest`.

To create a new release, just create a new tag off master.


## Contributing

We welcome pull requests. Please raise an issue to discuss your changes before
submitting a patch.


## Author

Vaidas Jablonskis [(vaijab)](https://github.com/vaijab)

