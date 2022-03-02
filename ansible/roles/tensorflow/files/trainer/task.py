import os
import json
from math import ceil
from argparse import ArgumentParser

import tensorflow as tf

from trainer.data import (numeric_feature_descriptions, get_dataset,
                          get_steps_per_epoch)
from trainer.model import create_model
from trainer.utils import save_model, Mode


def main(args):
    strategy = tf.distribute.MultiWorkerMirroredStrategy()

    with strategy.scope():
        multi_worker_model = create_model(
            numeric_feature_descriptions.keys(),
            args.layer_sizes,
            'mse',
            tf.keras.optimizers.Adam(args.learning_rate),
        )

    training_dataset = get_dataset(
        os.path.join(args.data_base_path, 'train/*.tfrecord'),
        strategy,
        args.train_batch_size,
        Mode.TRAIN,
    )

    validation_dataset = get_dataset(
        os.path.join(args.data_base_path, 'val/*.tfrecord'),
        strategy,
        args.validation_batch_size,
        Mode.VALID,
    )

    steps_per_epoch = get_steps_per_epoch(
        strategy,
        args.training_examples,
        args.train_batch_size,
    )
    validation_steps = get_steps_per_epoch(
        strategy,
        args.validation_examples,
        args.validation_batch_size,
    )

    multi_worker_model.fit(
        training_dataset,
        epochs=args.epochs,
        steps_per_epoch=steps_per_epoch,
        validation_data=validation_dataset,
        validation_steps=validation_steps,
    )

    eval_dataset = get_dataset(
        os.path.join(args.data_base_path, 'test/*.tfrecord'),
        strategy,
        args.evaluation_batch_size,
        Mode.EVAL,
    )
    evaluation_steps = get_steps_per_epoch(
        strategy,
        args.evaluation_examples,
        args.evaluation_batch_size,
    )

    multi_worker_model.evaluate(
        eval_dataset,
        steps=evaluation_steps,
    )

    save_model(
        os.path.join(args.job_dir, 'saved_model'),
        multi_worker_model,
    )


if __name__ == "__main__":
    parser = ArgumentParser()

    parser.add_argument("--job-dir", required=True)
    parser.add_argument("--layer-sizes", nargs="+", type=int, required=True)
    parser.add_argument("--learning-rate", type=float, required=True)
    parser.add_argument('--epochs', type=int, required=True)
    parser.add_argument("--data-base-path", required=True)
    parser.add_argument('--training-examples', type=int, required=True)
    parser.add_argument('--validation-examples', type=int, required=True)
    parser.add_argument('--evaluation-examples', type=int, required=True)
    parser.add_argument('--train-batch-size', type=int, default=1000)
    parser.add_argument('--validation-batch-size', type=int, default=10000)
    parser.add_argument('--evaluation-batch-size', type=int, default=10000)

    main(parser.parse_args())
