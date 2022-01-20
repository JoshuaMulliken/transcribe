# Transcribe

This is a command line tool that transcribes audio files into text using Otter.ai.
It is designed to be simple and portable.

## Usage

Once installed you can use the tool by running the following command:
```bash
$ transcribe [options] <audio-file-path>
```

Options:

- `-u`: Username for <https://otter.ai>.
- `-p`: Password for <https://otter.ai>.
- `-w`: Writes the config to a file. (default: `false`)
- `-c`: Custom path to the config file. (default: `<default-user-data-dir>/settings.json`)
- `-h` or `--help`: Prints the help message.

Since otter.ai is a cloud service, you will need to provide your credentials in order to use it. The tool will request
them from the user if they are not provided on first launch.

