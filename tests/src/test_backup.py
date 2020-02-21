
import os
import pprint
import shutil
import subprocess
import time

from kubernetes import client
import pytest

from common import kubeclient, util

def timestamp():
    return int(time.time())

def log_state(namespace, resdata):
    # Capture the state before cleaning up resources. This will help in
    # debugging.
    print("Output of 'describe all'")
    subprocess.call("kubectl describe persistentvolume", shell=True)
    subprocess.call("kubectl -n {} describe all".format(namespace), shell=True)
    subprocess.call("kubectl -n {} describe backuplocation".format(namespace), shell=True)
    subprocess.call("kubectl -n {} describe metadatabackuppolicy".format(namespace), shell=True)
    subprocess.call("kubectl -n {} describe metadatabackuprecord".format(namespace), shell=True)
    subprocess.call("kubectl -n {} describe metadatarestore".format(namespace), shell=True)
    subprocess.call("kubectl -n {} describe persistentvolumeclaim".format(namespace), shell=True)

    print("Output of 'logs'")
    for pod_name in resdata["pods"]:
        print("Output of 'logs' for {}".format(pod_name))
        subprocess.call("kubectl -n {} logs --all-containers {}".format(namespace, pod_name), shell=True)

    if "pv_path" in resdata:
        print("contents of PV dir {}".format(resdata["pv_path"]))
        cmd = "ls -lR {}".format(resdata["pv_path"])
        subprocess.call(cmd, shell=True)

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

    resdata = {"backuploc_creds": backuploc_creds, "pods": [], "backup_names": []}

    # If we create multiple resources in set up, we need to take care to do clean
    # up in case there are any errors. This is not an issue right now as we create
    # only one resource.

    yield resdata

    util.ignore_errors(lambda: log_state(globalconfig.namespace, resdata))

    util.ignore_errors_pred("restore_name" in resdata, lambda: globalconfig.mr_api.delete(resdata["restore_name"]))

    for backup_name in resdata.get("backup_names", []):
        util.ignore_errors(lambda: globalconfig.mbp_api.delete(backup_name))

    util.ignore_errors_pred("etcd_creds" in resdata, lambda: globalconfig.secret_api.delete(resdata["etcd_creds"]))
    util.ignore_errors_pred("backuploc_name" in resdata, lambda: globalconfig.backuploc_api.delete(resdata["backuploc_name"]))
    util.ignore_errors_pred("pvc_name" in resdata, lambda: globalconfig.pvc_api.delete(resdata["pvc_name"]))

    # PV should have been automatically deleted when PVC is deleted but just in case,
    # PVC was not created or to take care of any corner cases, try to delete pV any way.
    util.ignore_errors_pred("pv_name" in resdata, lambda: globalconfig.pvc_api.delete(resdata["pv_name"]))
    util.ignore_errors_pred("pv_path" in resdata, lambda: shutil.rmtree(resdata["pv_path"]))

    util.ignore_errors(lambda: globalconfig.secret_api.delete(backuploc_creds))

@pytest.mark.dependency()
def test_creating_backuplocation(globalconfig, resources):
    init_annotation = "initialized.annotations.kubedr.catalogicsoftware.com"
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

    assert len(pods.items) == 1, "Found pods: ({})".format(", ".join([x.metadata.name for x in pods.items]))
    pod_name = pods.items[0].metadata.name

    pod = kubeclient.wait_for_pod_to_be_done(pod_name)
    resources["pods"].append(pod_name)
    assert pod.status.phase == "Succeeded"

    backup_loc = globalconfig.backuploc_api.get(backuploc_name)
    assert backup_loc

    assert backup_loc["metadata"]["annotations"][init_annotation] == "true"

def do_backup(globalconfig, resources, backup_name, backup_spec):
    print("creating backup: {}".format(backup_name))
    globalconfig.mbp_api.create(backup_name, backup_spec)
    resources["backup_names"].append(backup_name)

    # Wait for cronjob to appear
    label_selector='kubedr.type=backup,kubedr.backup-policy={}'.format(backup_name)
    cronjobs = kubeclient.wait_for_cronjob_to_appear(label_selector)

    assert len(cronjobs.items) == 1
    cronjob_name = cronjobs.items[0].metadata.name

    # Wait for a backup pod to appear and then check its status.
    # Since the backup schedule is every minute, wait for slightly
    # longer than a minute before timing out.
    backup_pod = globalconfig.pod_api.get_by_watch(label_selector, timeout_seconds=75)

    pod_name = backup_pod.metadata.name
    resources["pods"].append(pod_name)

    phase = backup_pod.status.phase
    if phase == "Running" or phase == "Pending":
        pod = kubeclient.wait_for_pod_to_be_done(pod_name)
        backup_pod = globalconfig.pod_api.read(pod_name)

    assert backup_pod.status.phase == "Succeeded"
    policy = globalconfig.mbp_api.get(backup_name)
    pprint.pprint(policy)

    return policy

@pytest.mark.dependency(depends=["test_creating_backuplocation"])
def test_backup_without_certificates(globalconfig, resources):
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

    policy = do_backup(globalconfig, resources, backup_name, backup_spec)

    status = policy["status"]
    files_total = status["filesChanged"] + status["filesNew"]
    assert files_total == 1

@pytest.mark.dependency(depends=["test_creating_backuplocation"])
def test_backup_with_certificates(globalconfig, resources):
    if "etcd_data" not in globalconfig.testenv:
        pytest.skip("etcd data is not given, skipping...")

    if "certs_dir" not in globalconfig.testenv:
        pytest.skip("Certificates dir is not given, skipping...")

    etcd_data = globalconfig.testenv["etcd_data"]
    etcd_creds = "{}-{}".format("etcd-creds", timestamp())
    kubeclient.create_etcd_creds(etcd_creds, etcd_data["ca.crt"], etcd_data["client.crt"],
                                      etcd_data["client.key"])

    resources["etcd_creds"] = etcd_creds

    backup_name = "{}-{}".format("backup", timestamp())
    backup_spec = {
        "destination": resources["backuploc_name"],
        "certsDir": globalconfig.testenv["certs_dir"],
        "etcdCreds": etcd_creds,
        "schedule": "*/1 * * * *"
    }

    policy = do_backup(globalconfig, resources, backup_name, backup_spec)

    status = policy["status"]
    resources["mbr_with_certs"] = status["mbrName"]
    files_total = status["filesChanged"] + status["filesNew"]
    assert files_total > 1

@pytest.mark.dependency(depends=["test_backup_with_certificates"])
def test_restore(globalconfig, resources):
    pv = util.create_hostpath_pv()
    resources["pv_name"] = pv.metadata.name
    resources["pv_path"] = pv.spec.host_path.path

    pvc = util.create_pvc_for_pv(pv)
    resources["pvc_name"] = pvc.metadata.name

    mr_name = "{}-{}".format("mr", timestamp())
    mr_spec = {
        "mbrName": resources["mbr_with_certs"],
        "pvcName": resources["pvc_name"]
    }

    globalconfig.mr_api.create(mr_name, mr_spec)
    resources["restore_name"] = mr_name

    label_selector='kubedr.type=restore,kubedr.restore-mbr={}'.format(mr_spec["mbrName"])
    restore_pod = globalconfig.pod_api.get_by_watch(label_selector)

    pod_name = restore_pod.metadata.name
    resources["pods"].append(pod_name)

    phase = restore_pod.status.phase
    if phase == "Running" or phase == "Pending":
        pod = kubeclient.wait_for_pod_to_be_done(pod_name)
        restore_pod = globalconfig.pod_api.read(pod_name)

    assert restore_pod.status.phase == "Succeeded"
    assert os.path.exists("{}/data/etcd-snapshot.db".format(resources["pv_path"]))
    assert os.path.exists("{}/data/certificates".format(resources["pv_path"]))
    assert os.listdir(resources["pv_path"])
