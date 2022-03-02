import argparse
import json
import os

import tensorflow as tf

from dist_tf.dataset import mnist_dataset_creator
from dist_tf.models import build_mnist_model


def main():
    args = parse_args()
    _check_requirements_import()

    tf_config = os.environ.get('TF_CONFIG')
    if tf_config is None:
        raise argparse.ArgumentError(
            None, message='TF_CONFIG environment not set'
        )
    # TODO: the tf config validity should be probably guaranteed by the callee
    tf_config = json.loads(tf_config)
    print("TF_CONFIG: " + json.dumps(tf_config))
    strategy = tf.distribute.experimental.MultiWorkerMirroredStrategy()

    num_workers = len(tf_config['cluster']['worker'])
    global_batch_size = args.batch_size * num_workers

    with strategy.scope():
        dataset = strategy.experimental_distribute_datasets_from_function(
            lambda input_context: mnist_dataset_creator(
                global_batch_size, input_context
            )
        )
        model = build_mnist_model()

    model.fit(dataset, epochs=args.epochs, steps_per_epoch=70)


def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument('-b', '--batch-size', type=int, default=64)
    parser.add_argument('-e', '--epochs', type=int, default=5)

    return parser.parse_args()


def _check_requirements_import():
    import pandas as pd


if __name__ == '__main__':
    main()
