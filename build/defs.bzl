load("//build:bundle.bzl", _containers_bundle = "containers_bundle")
load("//build:image.bzl", _go_image = "go_image")
load("//build:push.bzl", _containers_push = "containers_push")

containers_bundle = _containers_bundle
containers_push = _containers_push
go_image = _go_image
