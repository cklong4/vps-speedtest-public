import pathlib,subprocess,tempfile,unittest
ROOT=pathlib.Path(__file__).resolve().parents[1]
class LauncherTests(unittest.TestCase):
 def test_shell_syntax_and_never_installs_python(self):
  script=(ROOT/'install.sh').read_text()
  subprocess.run(['sh','-n',str(ROOT/'install.sh')],check=True)
  self.assertNotIn('install -y python',script)
  self.assertIn('sha256sum -c',script)
  self.assertNotIn('--insecure',script)
 def test_release_manifest_has_exact_basename_and_hash(self):
  import hashlib
  rows=(ROOT/'release'/'SHA256SUMS').read_text().splitlines()
  for line in rows:
   digest,name=line.split()
   self.assertNotIn('/',name,'启动器按basename匹配，不能把release/写进发布校验表')
   self.assertEqual(digest,hashlib.sha256((ROOT/'release'/name).read_bytes()).hexdigest())
 def test_unknown_arch_exits_before_download(self):
  with tempfile.TemporaryDirectory() as d:
   p=pathlib.Path(d)/'uname';p.write_text('#!/bin/sh\ncase "$1" in -s) printf "Linux\\n";; -m) printf "mips64\\n";; esac\n');p.chmod(0o755)
   import os
   r=subprocess.run(['sh',str(ROOT/'install.sh')],env=dict(os.environ,PATH=d+':/usr/bin:/bin'),capture_output=True,text=True)
   self.assertEqual(r.returncode,2);self.assertIn('不支持',r.stdout)
if __name__=='__main__':unittest.main()
