from math import ceil
from enum import Enum, unique

import tensorflow as tf

from trainer.utils import Mode

numeric_feature_descriptions = {
    'feature1': tf.io.FixedLenFeature([], tf.float32),
    'feature2': tf.io.FixedLenFeature([], tf.float32),
    'feature3': tf.io.FixedLenFeature([], tf.float32),
    'feature4': tf.io.FixedLenFeature([], tf.float32),
    'feature5': tf.io.FixedLenFeature([], tf.float32),
    'feature6': tf.io.FixedLenFeature([], tf.float32),
    'feature7': tf.io.FixedLenFeature([], tf.float32),
    'feature8': tf.io.FixedLenFeature([], tf.float32),
    'feature9': tf.io.FixedLenFeature([], tf.float32)
}

target_description = {'target': tf.io.FixedLenFeature([], tf.float32)}


def _parse_example(example_proto):
    parsed_example = tf.io.parse_example(example_proto, {
        **numeric_feature_descriptions,
        **target_description
    })
    label = parsed_example.pop('target')
    return parsed_example, label


def _get_global_batch_size(strategy, batch_size):
    return strategy.num_replicas_in_sync * batch_size


def get_dataset(files_pattern, strategy, batch_size, mode):
    global_batch_size = _get_global_batch_size(strategy, batch_size)
    dataset = tf.data.TFRecordDataset(tf.io.gfile.glob(files_pattern))
    if mode != Mode.EVAL:
        dataset = dataset.cache()
    if mode == Mode.TRAIN:
        dataset = dataset.shuffle(10 * batch_size)
    dataset = dataset.repeat()
    dataset = dataset.batch(global_batch_size)
    dataset = dataset.map(
        _parse_example)  # vectorised tf.train.Example parsing
    dataset = dataset.prefetch(buffer_size=tf.data.experimental.AUTOTUNE)

    return dataset


def get_steps_per_epoch(strategy, nr_of_examples, batch_size):
    return ceil(nr_of_examples / _get_global_batch_size(strategy, batch_size))
