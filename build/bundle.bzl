load("@rules_multirun//:defs.bzl", _command = "command", _multirun = "multirun")
load("@rules_oci//oci:defs.bzl", _oci_load = "oci_load")

def containers_bundle(name, registry, images):
    [
        _oci_load(
            name = "{}_load_{}_{}".format(name, v.split("/")[-1], i),
            image = "{}:oci".format(v),
            repo_tags = ["{}/{}:{}".format(registry, k, "latest")],
        )
        for i, (k, v) in enumerate(images.items())
    ]

    [
        _command(
            name = "{}_cmd_{}_{}".format(name, v.split("/")[-1], i),
            arguments = [],
            command = ":{}_load_{}_{}".format(name, v.split("/")[-1], i),
        )
        for i, (_, v) in enumerate(images.items())
    ]

    _multirun(
        name = name,
        commands = [
            "{}_cmd_{}_{}".format(name, v.split("/")[-1], i)
            for i, (_, v) in enumerate(images.items())
        ],
        jobs = 0,
    )
