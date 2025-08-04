load("@rules_multirun//:defs.bzl", _command = "command", _multirun = "multirun")
load("@rules_oci//oci:defs.bzl", _oci_push = "oci_push")

def containers_push(name, images, registry, remote_tags = ["latest"]):
    [
        _oci_push(
            name = "{}_push_{}_{}".format(name, v.split("/")[-1], i),
            image = "{}:oci".format(v),
            repository = "{}/{}".format(registry, k),
            remote_tags = remote_tags,
        )
        for i, (k, v) in enumerate(images.items())
    ]

    [
        _command(
            name = "{}_cmd_push_{}_{}".format(name, v.split("/")[-1], i),
            command = ":{}_push_{}_{}".format(name, v.split("/")[-1], i),
            arguments = [],
        )
        for i, (_, v) in enumerate(images.items())
    ]

    _multirun(
        name = name,
        commands = [
            "{}_cmd_push_{}_{}".format(name, v.split("/")[-1], i)
            for i, (_, v) in enumerate(images.items())
        ],
        jobs = 0,
    )
