#!/usr/bin/env python3
"""
End-to-end manual test for DCT-51: audio_url dataset schema field +
`aitaskbuilder batch preview` command.

Builds the CLI, creates a dataset with an audio_url field, uploads sample
data, creates a batch, adds instructions, sets up the batch, then previews it.

Usage:
    python3 scripts/manual-tests/test_audio_batch_preview.py [workspace_id]

Requires PROLIFIC_TOKEN / PROLIFIC_URL set in the environment for the target
API. Run from anywhere; paths are resolved relative to the repo root.
"""

import os
import re
import subprocess
import sys
import tempfile

REPO_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
CLI_BINARY = os.path.join(tempfile.gettempdir(), "prolific-cli")
DEFAULT_WORKSPACE_ID = "679271425fe00981084a5f58"  # DCT Workspace


def run(args, check=True):
    print(f"\n$ {' '.join(args)}")
    result = subprocess.run(args, capture_output=True, text=True)
    print(result.stdout)
    if result.stderr:
        print(result.stderr, file=sys.stderr)
    if check and result.returncode != 0:
        print(f"Command failed with exit code {result.returncode}", file=sys.stderr)
        sys.exit(result.returncode)
    return result.stdout


def extract_field(output, label):
    match = re.search(rf"^{re.escape(label)}:\s*(\S+)", output, re.MULTILINE)
    if not match:
        print(f"Could not find '{label}:' in output:\n{output}", file=sys.stderr)
        sys.exit(1)
    return match.group(1)


def build_cli():
    print("Building CLI...")
    run(["go", "build", "-o", CLI_BINARY, "."], check=True)


def main():
    os.chdir(REPO_ROOT)
    workspace_id = sys.argv[1] if len(sys.argv) > 1 else DEFAULT_WORKSPACE_ID

    build_cli()

    # 1. Create dataset with an audio_url field
    schema = (
        '{"fields":{"question":{"type":"text","label":"Question"},'
        '"clip":{"type":"audio_url","label":"Audio clip"}}}'
    )
    output = run(
        [
            CLI_BINARY, "aitaskbuilder", "dataset", "create",
            "-n", "Audio URL Test Dataset",
            "-w", workspace_id,
            "--strict",
            "--schema", schema,
        ]
    )
    dataset_id = extract_field(output, "ID")
    print(f"Created dataset: {dataset_id}")

    # 2. Upload sample data containing audio URLs
    csv_path = os.path.join(tempfile.gettempdir(), "audio-dataset.csv")
    with open(csv_path, "w") as f:
        f.write(
            "question,clip\n"
            '"What emotion is being expressed?","https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3"\n'
            '"Transcribe the spoken word.","https://www.w3schools.com/html/horse.mp3"\n'
        )

    dataset_id = "019f6123-1fdc-7518-a3ec-75b830c5a5f4"
    run([CLI_BINARY, "aitaskbuilder", "dataset", "upload", "-d", dataset_id, "-f", csv_path])

    # 3. Check dataset status (informational; may need polling until READY)
    run([CLI_BINARY, "aitaskbuilder", "dataset", "check", "-d", dataset_id], check=False)

    # 4. Create batch linked to the dataset
    output = run(
        [
            CLI_BINARY, "aitaskbuilder", "batch", "create",
            "-n", "Audio URL Test Batch",
            "-w", workspace_id,
            "-d", dataset_id,
            "--task-name", "Audio Review Task",
            "--task-introduction", "Listen to the audio clip and answer the question.",
            "--task-steps", "1. Listen to the clip\\n2. Answer the question",
        ]
    )
    batch_id = extract_field(output, "ID")
    print(f"Created batch: {batch_id}")

    # 5. Add instructions (required before setup)
    run(
        [
            CLI_BINARY, "aitaskbuilder", "batch", "instructions",
            "-b", batch_id,
            "-j", '[{"type":"free_text","created_by":"Cem","description":"Please describe what you heard."}]',
        ]
    )

    # 6. Setup the batch
    run(
        [
            CLI_BINARY, "aitaskbuilder", "batch", "setup",
            "-b", batch_id,
            "-d", dataset_id,
            "--tasks-per-group", "1",
        ]
    )

    # 7. Preview the batch (the new command under test)
    run([CLI_BINARY, "aitaskbuilder", "batch", "preview", "-b", batch_id])

    print("\nDone. Review output above for correctness.")


if __name__ == "__main__":
    main()
