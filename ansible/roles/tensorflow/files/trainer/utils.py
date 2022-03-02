import os
from time import sleep
from enum import Enum, unique

import tensorflow as tf


@unique
class Mode(Enum):
    TRAIN = 1
    VALID = 2
    EVAL = 3


def _is_chief(cluster_resolver):
    task_type = cluster_resolver.task_type
    return task_type is None or task_type == 'chief'


def _get_temp_dir(model_path, cluster_resolver):
    worker_temp = f'worker{cluster_resolver.task_id}_temp'
    return os.path.join(model_path, worker_temp)


def save_model(model_path, model):
    # the following is need for TF 2.2. 2.3 onward, it can be accessed from strategy
    cluster_resolver = tf.distribute.cluster_resolver.TFConfigClusterResolver()
    is_chief = _is_chief(cluster_resolver)

    if not is_chief:
        model_path = _get_temp_dir(model_path, cluster_resolver)

    model.save(model_path)

    if is_chief:
        # wait for workers to delete; check every 100ms
        # if chief is finished, the training is done
        while tf.io.gfile.glob(os.path.join(model_path, "worker*")):
            sleep(0.1)

    if not is_chief:
        tf.io.gfile.rmtree(model_path)
