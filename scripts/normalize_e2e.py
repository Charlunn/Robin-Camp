from pathlib import Path

root = Path(__file__).resolve().parents[1]
script_path = root / 'app' / 'e2e-test.sh'
data = script_path.read_bytes().replace(b'\r\n', b'\n').replace(b'\r', b'')
script_path.write_bytes(data)
print('normalized to LF, bytes:', len(data))
