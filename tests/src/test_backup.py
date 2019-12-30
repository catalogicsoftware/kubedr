
import pprint
import time

from kubernetes import client
import pytest

from common import kubeclient, util

def timestamp():
    return int(time.time())

# "resources" is used to store state as resources are being created.
# This allows us to delete all the resources in one place and also
# enables deletion even in case of test failures.
@pytest.fixture(scope="module")
def resources(globalconfig):
    if not globalconfig.testenv:
        pytest.skip("Test environment data is not given, skipping...")
    backuploc = globalconfig.testenv["backuploc"]

    backuploc_creds = "{}-{}".format("s3creds", timestamp())
    kubeclient.create_backuploc_creds(backuploc_creds, backuploc["access_key"], backuploc["secret_key"],
                                      globalconfig.restic_password)

    resdata = {"backuploc_creds": backuploc_creds}

    yield resdata

    if "backup_name" in resdata:
        util.ignore_errors(lambda: globalconfig.mbp_api.delete(resdata["backup_name"]))

    if "etcd_creds" in resdata:
        util.ignore_errors(lambda: globalconfig.secret_api.delete(resdata["etcd_creds"]))

    if "backuploc_name" in resdata:
        util.ignore_errors(lambda: globalconfig.backuploc_api.delete(resdata["backuploc_name"]))

    util.ignore_errors(lambda: globalconfig.secret_api.delete(backuploc_creds))

@pytest.mark.dependency()
def test_creating_backuplocation(globalconfig, resources):
    endpoint = globalconfig.testenv["backuploc"]["endpoint"]

    bucket_name = "{}-{}".format(
        globalconfig.testenv["backuploc"]["bucket_name_prefix"],
        timestamp())

    backuploc_name = "{}-{}".format("tests3", timestamp())
    backuploc_spec = {
        "url": endpoint,
        "bucketName": bucket_name,
        "credentials": resources["backuploc_creds"]
    }
    globalconfig.backuploc_api.create(backuploc_name, backuploc_spec)
    resources["backuploc_name"] = backuploc_name

    label_selector='kubedr.type=backuploc-init,kubedr.backuploc={}'.format(backuploc_name)
    pods = kubeclient.wait_for_pod_to_appear(label_selector)

    assert len(pods.items) == 1
    pod_name = pods.items[0].metadata.name

    pod = kubeclient.wait_for_pod_to_be_done(pod_name)
    assert pod.status.phase == "Succeeded"

@pytest.mark.dependency(depends=["test_creating_backuplocation"])
def test_backup(globalconfig, resources):
    if "etcd_data" not in globalconfig.testenv:
        pytest.skip("etcd data is not given, skipping...")

    etcd_data = globalconfig.testenv["etcd_data"]
    etcd_creds = "{}-{}".format("etcd-creds", timestamp())
    kubeclient.create_etcd_creds(etcd_creds, etcd_data["ca.crt"], etcd_data["client.crt"],
                                      etcd_data["client.key"])

    resources["etcd_creds"] = etcd_creds

    backup_name = "{}-{}".format("backup", timestamp())
    backup_spec = {
        "destination": resources["backuploc_name"],
        "etcdCreds": etcd_creds,
        "schedule": "*/1 * * * *"
    }

    print("creating backup: {}".format(backup_name))
    globalconfig.mbp_api.create(backup_name, backup_spec)
    resources["backup_name"] = backup_name

    # Wait for cronjob to appear
    label_selector='kubedr.type=backup,kubedr.backup-policy={}'.format(backup_name)
    cronjobs = kubeclient.wait_for_cronjob_to_appear(label_selector)

    assert len(cronjobs.items) == 1
    cronjob_name = cronjobs.items[0].metadata.name

    # Wait for a backup pod to appear and then check its status.
    # Since the backup schedule is every minute, wait for slightly
    # longer than a minute before checking.
    time.sleep(70)

    pods = kubeclient.wait_for_pod_to_appear(label_selector)

    assert len(pods.items) == 1
    pod_name = pods.items[0].metadata.name

    pod = kubeclient.wait_for_pod_to_be_done(pod_name)
    assert pod.status.phase == "Succeeded"
