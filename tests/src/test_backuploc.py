
import pprint
import subprocess
import time

from kubernetes import client
import pytest

from common import kubeclient, util

def timestamp():
    return int(time.time())

# We don't want to test with incorrect IP as repo init will take more than
# a minute to time out, thus adding to test time.
def test_creating_backuplocation_with_invalid_credentials(globalconfig):
    if not globalconfig.testenv:
        pytest.skip("Test environment data is not given, skipping...")

    backuploc_creds_created = False
    backuploc_created = False
    init_annotation = "initialized.annotations.kubedr.catalogicsoftware.com"
    backuploc_creds = "{}-{}".format("s3creds", timestamp())
    backuploc_name = "{}-{}".format("tests3", timestamp())
    backup_loc = None

    backuploc = globalconfig.testenv["backuploc"]
    endpoint = globalconfig.testenv["backuploc"]["endpoint"]

    bucket_name = "{}-{}".format(
        globalconfig.testenv["backuploc"]["bucket_name_prefix"],
        timestamp())

    backuploc_spec = {
        "url": endpoint,
        "bucketName": bucket_name,
        "credentials": backuploc_creds
    }

    try:
        kubeclient.create_backuploc_creds(backuploc_creds, backuploc["access_key"], backuploc["secret_key"]+"s",
                                          globalconfig.restic_password)
        backuploc_creds_created = True

        globalconfig.backuploc_api.create(backuploc_name, backuploc_spec)
        backuploc_created = True

        label_selector='kubedr.type=backuploc-init,kubedr.backuploc={}'.format(backuploc_name)
        pods = kubeclient.wait_for_pod_to_appear(label_selector)

        assert len(pods.items) == 1, "Found pods: ({})".format(", ".join([x.metadata.name for x in pods.items]))
        pod_name = pods.items[0].metadata.name

        pod = kubeclient.wait_for_pod_to_be_done(pod_name)

        backup_loc = globalconfig.backuploc_api.get(backuploc_name)
        assert backup_loc

        # We expect backup location initialization to fail so init annotation
        # should not be set.
        assert ("annotations" not in backup_loc["metadata"] or 
                init_annotation not in backup_loc["metadata"]["annotations"] or 
                backup_loc["metadata"]["annotations"][init_annotation] == "false")
    finally:
        if backup_loc:
            pprint.pprint(backup_loc)
        util.ignore_errors_pred(backuploc_created, lambda: globalconfig.backuploc_api.delete(backuploc_name))
        util.ignore_errors_pred(backuploc_creds_created, lambda: globalconfig.secret_api.delete(backuploc_creds))

