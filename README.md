# Docker instance-principal credential helper for OCIR

Use the OCI go sdk to request a docker token based on the OCI instance principal.

# Use

Build the helper:

    go mod vendor
    go build docker-credential-ocir.go

Install it somewhere on your path:

    sudo cp docker-credential-ocir /usr/local/bin

Configure your local docker installation to use the credential helper:

    mkdir -p ~/.docker
    cat > ~/.docker/config.json <<EOF
    {
        "credsStore": "ocir"
    }
    EOF

There are details on constructing a more nuanced configuration on
[the docker website](https://docs.docker.com/engine/reference/commandline/login/)

# OCI policy configuration

Construct a dynamic-group definition that includes your instance:

    # Dynamic group `example-instance-dynamic-group`
    instance.compartment.id = 'ocid1.compartment.oc1..aaaaaaaawflibbertigibbetblahblahblahblah'

Construct a policy that permits the instance the rights you want:

    # Root policy `example-instance-repo-management`
    allow dynamic-group example-instance-dynamic-group to manage repos in tenancy where all {target.repo.name = /example*/}

# Try it

On the instance:

    docker pull iad.ocir.io/blahblah/example/repo/path:0.0.1

