# confv

confv is a Kubernetes [flexvolume](https://github.com/kubernetes/community/blob/master/contributors/devel/flexvolume.md)
driver which aims to dynamically render application config according to each instance's
identity.

# Building

You need [Go](https://golang.org/dl/) and [Docker](https://www.docker.com/get-started)
installed.

```bash
$ go get github.com/dunjut/confv
$ cd $GOPATH/src/github.com/dunjut/confv
$ make image
```

You should now have two new docker images on your machine:

```bash
$ docker images
REPOSITORY              TAG              IMAGE ID              CREATED              SIZE
dunjut/confv-install    latest           08454fbd4870          5 seconds ago        72.7MB
dunjut/kubectl          v1.10.0          ded148a7f3d8          40 seconds ago       113MB
```

**NOTE:** users can directly use these images from docker hub. Current uploaded:

- dunjut/confv-install:0.1.0
- dunjut/kubectl:v1.10.0

# Usage

## Cluster administrator

```bash
$ kubectl apply -f install/confv.yaml
```

This will install the confv binary on each of your Kubernetes node's flexvolume plugin
directory and start `kubectl proxy` daemon on each node.

## Normal user

1. create a [ConfigMap](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/) 
using Go's [template syntax](https://golang.org/pkg/text/template/). For example:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: myconfig
data:
  tpl: |-
    {
      "role": "{{ .role }}",
      {{- if .backup.enabled }}
      "backupStorage": "{{ .backup.storage }}",
      {{- end }}
      "bindAddr": ":9999"
    }
```

Here our exmaple config is in json format, and we've left some variables using Go's
template syntax. Those variables are defined in another place, see next section.

2. create another ConfigMap to hold config values. For example:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: myvalues
data:
  database-0: |-
    role: master
    backup:
      enabled: false
  database-1: |-
    role: slave
    backup:
      enabled: true
      storage: aws-s3
```

Here we provide two instances' config values, one is `database-0` and the other is
`database-1`. You should have noticed that their config values are different.

**Note:** config values MUST be in yaml format, otherwise confv won't know how to map
these values to corresponding variables.

3. create workload resource using Deployment/DaemonSet/etc. For example:

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: database
spec:
  serviceName: database
  replicas: 2
  selector:
    matchLabels:
      app: database
  template:
    metadata:
      labels:
        app: database
    spec:
      containers:
      - name: database
        image: docker.io/citizenstig/httpbin
        volumeMounts:
        - name: config
          mountPath: /app
      volumes:
      - name: config
        flexVolume:
          driver: "dunjut/confv"
          options:
            template: "name=myconfig,key=tpl"
            values: "name=myvalues,identifiedBy=podName"
            target: "db.cnf"
```

In this example, we mount our application config using a flexVolume. The driver here is
`dunjut/confv` and the options are:

**template**: specify where your config template lives, expected to be a ConfigMap.
- name: name of the ConfigMap resource (namespace expected to be the same as Pod)
- key: since a ConfigMap may contain multiple kv pairs, a key need to be specified.

**values**: specify where your config values lives, expected to be a ConfigMap.
- name: name of the ConfigMap resource (namespace expected to be the same as Pod)
- identifiedBy: indicates how to define the application identity of each Pod, must be one
of `hostIP`, `nodeName` or `podName` (In the example we're using podName since it's
predictable in a StatefulSet). This field will later be used as a key when getting values
in the values ConfigMap.

**target**: specify what filename to use when writing rendered config.

## Test

After all of the above steps, you should get your workloads running. In our example, you
could expect the following rendered results:

```bash
$ kubectl exec database-0 cat /app/db.cnf
{
  "role": "master",
  "bindAddr": ":9999"
}

$ kubectl exec database-1 cat /app/db.cnf
{
  "role": "slave",
  "backupStorage": "aws-s3",
  "bindAddr": ":9999"
}
```

# Contact

Feel free to send me feedbacks via email or issues. Pull requests are welcomed!

