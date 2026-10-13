#!/usr/bin/env python3
"""Check pins, workspace tags and the ECS -> GUI dependency contract."""
import argparse
import json
from pathlib import Path
import re
import subprocess

PREFIX = 'github.com/oneclickvirt/'

def modules(folder):
    value = json.loads(subprocess.check_output(['go', 'mod', 'edit', '-json'], cwd=folder, text=True))
    return {r['Path']:r['Version'] for r in value.get('Require', [])}

def field(path, name):
    match = re.search(r'\b' + name + r'\s*=\s*"([^"]+)"', path.read_text())
    if not match:
        raise RuntimeError('missing metadata')
    return match[1]

def verify(workspace=False):
    pinned = modules('.')
    upstream = pinned[PREFIX+'ecs']
    version = Path('VERSION').read_text().strip()
    if not re.fullmatch(r'\d+\.\d+\.\d+', version):
        raise RuntimeError('invalid application version')
    if field(Path('FyneApp.toml'), 'Version') != version or field(Path('internal/appmeta/version.go'), 'Version') != version:
        raise RuntimeError('GUI metadata version mismatch')
    if field(Path('internal/appmeta/version.go'), 'UpstreamECSVersion') != upstream:
        raise RuntimeError('ECS metadata version mismatch')
    if not workspace:
        return version, upstream
    required = modules('deps/ecs')
    for module, ref in required.items():
        if module.startswith(PREFIX) and pinned.get(module) != ref:
            raise RuntimeError('GUI component graph differs from released ECS')
    if field(Path('deps/ecs/api/api.go'), 'DefaultVersion') != upstream:
        raise RuntimeError('checked out ECS source differs from pinned version')
    for module, ref in pinned.items():
        if not module.startswith(PREFIX):
            continue
        source = Path('deps')/module[len(PREFIX):]
        if not source.exists():
            continue
        actual = subprocess.check_output(['git','rev-parse','HEAD'],cwd=source,text=True).strip()
        expected = subprocess.check_output(['git','rev-parse',ref+'^{}'],cwd=source,text=True).strip()
        if actual != expected:
            raise RuntimeError('component checkout is not the immutable pinned tag')
    return version, upstream

if __name__ == '__main__':
    p = argparse.ArgumentParser()
    p.add_argument('--workspace', action='store_true')
    p.add_argument('--outputs')
    args = p.parse_args()
    version, upstream = verify(args.workspace)
    if args.outputs:
        with open(args.outputs, 'a') as output:
            output.write('GUI_VERSION='+version+'\nECS_VERSION='+upstream[1:]+'\n')
