from pathlib import Path
p = Path('cinema/e2e-test.sh')
data = p.read_bytes().replace(b'\r\n', b'\n').replace(b'\r', b'')
p.write_bytes(data)
print('normalized to LF, bytes:', len(data))
