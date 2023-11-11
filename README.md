# Rebase Migration

## Overview

The Rebase Migration script is a tool designed to manage SQL migration file versions within a Git repository when merging feature branches into the master branch. Despite its name, the script does not perform a `git rebase` but instead mirrors the behavior of rebasing by ensuring that migration scripts are sequentially versioned and that any conflicts are resolved in alignment with the merge order into the master branch.

## Clarification on Naming

The term "rebase" in the context of this script does not refer to the Git operation `git rebase`. Instead, it reflects the concept of re-establishing a base for migration versions — similar to how `git rebase` re-establishes a base for a series of commits. This script ensures that your migration files' versioning is re-based according to the order in which they are merged into the master branch, thus preventing conflicts and maintaining the integrity of your migration sequence.

## Why We Don't Use Timestamps

We avoid using timestamps for versioning because they do not accurately represent the sequential order of migrations within the master branch. The script gives precedence to the migration that is merged into the master branch first, ensuring a consistent and conflict-free version history, irrespective of when migrations were originally created in feature branches.

## Installation

1. Clone the repository with the migration script to your local machine.
2. Ensure Go is installed on your system.
3. Compile the script: `go build rebase-migration.go`.

## Usage

Run this script from the feature branch you intend to merge:

```sh
./rebase-migration <path/to/migration>
```

Replace `<path/to/migration>` with the actual path to your migration directory.

## How It Works

The script executes the following:

1. Scans for migration files that could cause versioning conflicts during a merge.
2. Orders migrations based on commit history to determine the correct versioning.
3. Assigns new version numbers to conflicting migration files, respecting their merge order.
4. Stages and commits the updated files to the feature branch.

## Integration with CI/CD

Designed for CI/CD integration, this script can be included in workflows, such as GitHub Actions, to automatically resolve migration conflicts as part of the merge process.

## Best Practices

- Backup migration files before running the script.
- Test the script in a controlled environment.
- Communicate with your team to ensure everyone is informed about the script's usage and the migration process.

## Contributing

Contributions are welcome. Please ensure changes are well-tested before creating a pull request.
