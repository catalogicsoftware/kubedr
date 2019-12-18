
import time

from kubernetes import client
import pytest

from common import kubeclient

def timestamp():
    return int(time.time())

@pytest.fixture()
def resources(globalconfig):
    if not globalconfig.testenv:
        pytest.skip("Test environment data is not given, so skipping...")
    backuploc = globalconfig.testenv["backuploc"]

    creds_name = "{}-{}".format("s3creds", timestamp())
    kubeclient.create_backuploc_creds(creds_name, backuploc["access_key"], backuploc["secret_key"],
                                      backuploc["restic_password"])

    resdata = {"creds_name": creds_name}

    yield resdata

    if "backuploc_name" in resdata:
        globalconfig.backuploc_api.delete(resdata["backuploc_name"])

    globalconfig.secret_api.delete(creds_name)

def test_creating_backuplocation(globalconfig, resources):
    endpoint = globalconfig.testenv["backuploc"]["endpoint"]
    bucket_name = "testbucket-{}".format(timestamp())

    backuploc_name = "{}-{}".format("tests3", timestamp())
    backuploc_spec = {
        "url": endpoint,
        "bucketName": bucket_name,
        "credentials": resources["creds_name"]
    }
    globalconfig.backuploc_api.create(backuploc_name, backuploc_spec)
    resources["backuploc_name"] = backuploc_name

    label_selector='kubedr.type=backuploc-init,kubedr.backuploc={}'.format(backuploc_name)
    pods = kubeclient.wait_for_pod_to_appear(label_selector)

    assert len(pods.items) == 1
    pod_name = pods.items[0].metadata.name

    pod = kubeclient.wait_for_pod_to_be_done(pod_name)
    assert pod.status.phase == "Succeeded"
