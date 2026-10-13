"""Keep public GUI source pinned to the tested upstream, including cold caches."""

import importlib.util
import json
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

script = Path(__file__).resolve().parents[2] / ".back/create_public_branch.py"
specification = importlib.util.spec_from_file_location("public_source", script)
generator = importlib.util.module_from_spec(specification)
specification.loader.exec_module(generator)


class PublicSourceResolutionTest(unittest.TestCase):
    def resolve(self, actual="v0.2.14", requirements="module github.com/oneclickvirt/ecs\n"):
        with tempfile.TemporaryDirectory() as temporary:
            source = Path(temporary) / "source"
            (source / "api").mkdir(parents=True)
            (source / "api/api.go").write_text(f'var DefaultVersion = "{actual}"\n')
            (source / "go.mod").write_text(requirements)
            responses = [
                json.dumps({"Path": "github.com/oneclickvirt/ecs", "Version": "v0.2.15-0.20261013112601-aabbccddeeff"}),
                json.dumps({"Dir": str(source)}), "", "",
            ]
            with patch.object(generator, "read_file", return_value='var UpstreamECSVersion = "v0.2.14"'), patch.object(generator, "run_go", side_effect=responses) as command:
                try:
                    generator.use_public_goecs()
                except ValueError:
                    return command.call_args_list, True
                return command.call_args_list, False

    def test_resolve_immutable_source_before_loading_consumer_graph(self):
        commands, failed = self.resolve()
        self.assertFalse(failed)
        self.assertEqual(commands[0].args, ("list", "-m", "-json", "github.com/oneclickvirt/ecs@public-v0.2.14"))
        self.assertIsNotNone(commands[0].kwargs["cwd"])
        self.assertEqual(commands[2].args[:2], ("mod", "edit"))
        self.assertEqual(commands[3].args, ("mod", "tidy"))

    def test_reject_wrong_public_source_version(self):
        commands, failed = self.resolve(actual="v0.2.15")
        self.assertTrue(failed)
        self.assertEqual(len(commands), 2)

    def test_reject_private_module_requirements(self):
        commands, failed = self.resolve(requirements="module github.com/oneclickvirt/ecs\nrequire github.com/oneclickvirt/security v0.0.21\n")
        self.assertTrue(failed)
        self.assertEqual(len(commands), 2)

    def test_missing_public_release_never_falls_back_to_mutable_branch(self):
        with patch.object(generator, "read_file", return_value='var UpstreamECSVersion = "v0.2.14"'), patch.object(generator, "run_go", side_effect=RuntimeError("unavailable")) as command:
            with self.assertRaises(RuntimeError):
                generator.use_public_goecs()
            self.assertEqual(command.call_count, 1)
            self.assertNotIn("@public", command.call_args.args[-1].replace("@public-v0.2.14", ""))

    def test_transport_timeouts_have_a_bounded_retry_budget(self):
        with patch.object(generator.subprocess, "run", side_effect=subprocess.TimeoutExpired("go", 180)) as command, patch.object(generator.time, "sleep"):
            with self.assertRaises(RuntimeError):
                generator.run_go("list", "-m", "example.invalid/module@v1.0.0")
            self.assertEqual(command.call_count, 3)
            self.assertTrue(all(call.kwargs["timeout"] == 180 for call in command.call_args_list))


if __name__ == "__main__":
    unittest.main()
