#!/usr/bin/env python
'''
test engine for jenkins
'''
import os
import sys
import subprocess

try:
    import pip
except ImportError:
    print "you must have pip installed."
    print "You can install it using https://bootstrap.pypa.io/get-pip.py"
    raise

try:
    import glob
except ImportError:
    pip.main(["install", "glob"])
try:
    from termcolor import colored
except ImportError:
    pip.main(["install", "termcolor"])
    from termcolor import colored
try:
    from jinja2 import Environment, FileSystemLoader
except ImportError:
    pip.main(["install", "jinja2"])
    from jinja2 import Environment, FileSystemLoader
try:
    import psutil
except ImportError:
    pip.main(["install", "psutil"])
    import psutil
try:
    import shutil
except ImportError:
    pip.main(["install", "shutil"])
    import shutil



GO_PATH = "/usr/local/go/bin"
# dirs without go tests
SKIP_TEST_LIST = ['/doc', '/tmp']

def copy_and_overwrite(from_path, to_path):
    if os.path.exists(to_path):
        shutil.rmtree(to_path)
    shutil.copytree(from_path, to_path)



class GoTests(object):
    '''go test framework for jenkins '''

    def __init__(self):
        '''
        create mian path, save cwd and cd to main project path
        '''
        self.workspace = os.environ.get("WORKSPACE")
        self.runBenchmark = False
        self.failed_tests = list()
        if self.workspace is None:
            self.current_path = os.getcwd()
            self._main_path = os.path.dirname(sys.argv[0])
            self.template_dir = os.path.join(
                os.path.dirname(os.path.abspath(__file__)), "tmp")
        else:
            # set jenkins env jenkins
            _path = os.environ.get("PATH")
            if self.workspace not in _path:
                _path += ":%s/bin" % self.workspace
            if GO_PATH not in _path:
                _path += ":%s" % GO_PATH
            os.environ["PATH"] = _path
            # now we need to copy to match go src standrts
            os.environ["GOPATH"] = self.workspace
            #create src
            go_src=os.path.join(self.workspace, "src")
            go_project_src=os.path.join(go_src ,"quickstep")
            if not os.path.exists(go_src):
                os.mkdir(go_src)
            if not os.path.exists(go_project_src) :
                os.mkdir(go_project_src)
	    _temp_path_src=os.path.join(self.workspace , "backend")
	    _temp_path_target=os.path.join(go_project_src , "backend")
	    copy_and_overwrite(_temp_path_src,_temp_path_target)
            #create new workspace
            self.workspace =  os.path.join(self.workspace, "src/quickstep")
            self._main_path = os.path.join(self.workspace, "backend")
            self.template_dir = os.path.join(self._main_path, "tmp")
	   
        os.chdir(self._main_path)
        # get all dependencies
        if not self.check_databases():
            print colored('ERROR: Database check failed', 'red')
            print "for installation refer to: "
            print "https://docs.mongodb.com/manual/administration/install-on-linux/"
            print
	    raise TypeError
        else:
            print "Database : ", colored("OK", "green")

        if not self.load_imports():
            print colored('ERROR: Import failed', 'red')
        else:
            print "Import: ", colored("OK", "green")

    @classmethod
    def _get_all_dirs(cls, path, skip=None):
        '''
        get all available directories
        @in path (string)
        @out list of dirs ( strings)
        '''
        all_dirs = []
        if skip is None:
            skip = []
        for root, dirs, _ in os.walk(path):
            for name in dirs:
                skip_entry = False
                entry = os.path.join(root, name)
                if len(skip):
                    for item in skip:
                        if item in entry:
                            skip_entry = True
                            break
                if not skip_entry:
                    all_dirs.append(entry)

        return all_dirs

    @classmethod
    def _check_one_dir_for_test(cls, path):
        '''
        return number of test files  existsing  in path
        @in path (string)
        @out nr of test files (int)
        '''
        return len(glob.glob('%s/*_test.go' % path))

    @classmethod
    def progress(cls, count, total, suffix='', prefix=''):
        '''
        progress bar
        '''
        bar_len = 60
        filled_len = int(round(bar_len * count / float(total)))

        percents = round(100.0 * count / float(total), 1)
        dbar = '=' * filled_len + '-' * (bar_len - filled_len)

        sys.stdout.write('%s[%s] %s%s ...%s\r' %
                         (prefix, dbar, percents, '%', suffix))
        sys.stdout.flush()

    def _parse_line(self, line, sub_path):
        ''' parse one line from xml '''
        _errfound = False
        item_dict = line.split(" ")
        for item in item_dict:
            if item.startswith("errors="):
                if item.split("=")[1] != '"0"':
                    self.failed_tests.append(
                        "%s#%s" % (sub_path, "have errors"))
                    _errfound = True
            if item.startswith("failures="):
                if item.split("=")[1] != '"0"':
                    self.failed_tests.append("%s#%s" % (sub_path, "failed"))
                    _errfound = True
        return _errfound

    def check_and_generate_sys_error(self, sub_path, test_path):
        ''' generate error for if we can get it using go2xunit
         ex: compilation error
         '''
        if os.stat(test_path).st_size > 0:
            if self.workspace is None:
                _file = open(test_path)
                for line in _file.readlines():
                    if self._parse_line(line, sub_path):
                        break
                _file.close()

            return
        args = "go test -v | go2xunit"
        process = subprocess.Popen(
            args, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        process.wait()
        err_output = process.communicate()
        template_file = "err_template.txt"
        failure = Environment(
            loader=FileSystemLoader(self.template_dir), trim_blocks=True
        ).get_template(template_file).render(
            testname="OS_SystemError",
            subpath=sub_path,
            error="\n".join(err_output)
        )
        _file = open(test_path, 'w')
        _file.write(failure)
        _file.close()
        if self.workspace is None:
            self.failed_tests.append("%s#%s" % (sub_path, "system error"))
        return

    def run(self):
        '''
        run all go tests and convert them to format
        which is understandable for jenkins
        '''
        dirs = self._get_all_dirs(self._main_path, skip=SKIP_TEST_LIST)
        count = 0
        resp = 0
        total = len(dirs) - 1
        zero_length_tests = []
        if self.workspace  is not None:
            print "\n===== GO ======"
        for subdir in dirs:
            nr_of_tests = self._check_one_dir_for_test(subdir)
            if not nr_of_tests:
                zero_length_tests.append(subdir)
            else:
                _my_current_path = os.getcwd()
                os.chdir(subdir)
                test_output = 'T'.join(subdir.strip(".").split("/")) + ".xml"
                bench_output = 'B'.join(
                    subdir.strip(".").split("/")) + ".banch"
                coverage_output = 'C'.join(
                    subdir.strip(".").split("/")) + ".cub"
                if self.workspace is not None:
                    print "test: ", colored(subdir, "blue")
                else:
                    self.progress(count, total, "complete",
                                  "go test %4d :  " % count)
                try:
                    # run short
                    args = "go test >/dev/null 2>>/dev/null"
                    prc = subprocess.Popen(args, shell=True)
                    prc.wait()
                    resp = prc.returncode

                    # run loing and prse
                    if self.workspace is not None:
                        args = "go test -v | go2xunit > %s" % test_output
                    else:
                        args = "go test -v 2>&1 | go2xunit > %s 2>>/dev/null" % test_output
                    prc = subprocess.Popen(args, shell=True)
                    prc.wait()
                    # check if test_outpu failed
                    self.check_and_generate_sys_error(subdir, test_output)

                    # if short succeed
                    # run coverage
                    if not resp:
                        args = "gocov test ./ 2>>/dev/null | gocov-xml > %s 2>>/dev/null" % coverage_output
                        prc = subprocess.Popen(args, shell=True)
                        prc.wait()
                        if self.runBenchmark:
                            # run benchmark
                            # go test -bench=.
                            args = "go test -bench=. | gobench2plot > %s 2>>/dev/null" % bench_output
                            prc = subprocess.Popen(args, shell=True)
                            prc.wait()
                finally:
                    os.chdir(_my_current_path)
            count += 1
        sys.stdout.write("\n")
        sys.stdout.flush()
        if self.workspace is None and len(self.failed_tests):
            print colored("\nTEST RESULTS:\n================", "blue")
            for f_test in self.failed_tests:
                msg = f_test.split('#')
                print "test in: ", colored(msg[0], attrs=['bold']), colored(msg[1], "red")
            print colored("================\n", "blue")
        else:
            if len(zero_length_tests):
                print "\n==== ERRORS ===="

        for subdir in zero_length_tests:
            print colored('ERROR:', 'red'), "subdirectory: ", colored(subdir, attrs=['bold']), "- have no tests defined !!!!"

    def check_databases(self):
	''' check if databases are running '''
        found = False
	for proc in psutil.process_iter():
	    try:
                pinfo = proc.as_dict(attrs=['pid', 'name'])
            except psutil.NoSuchProcess:
                pass
            else:
                if pinfo['name'] == 'mongod':
                    found = True
	return found

    def load_imports(self):
        ''' load all go imports '''
        try:
            _load_file = open(os.path.join(
                self._main_path, "tmp/go_import_extra.txt"))
        except IOError:
            return False
        prc = subprocess.Popen("go get .", shell=True)
        prc.wait()
        if prc.returncode != 0:
            return False
        resp = True
        for line in _load_file.readlines():
            prc = subprocess.Popen("go get %s" % line, shell=True)
            prc.wait()
            if prc.returncode != 0:
                resp = False
        _load_file.close()
        return resp

    @classmethod
    def build(cls):
        ''' build executables'''
        resp = True
        prc = subprocess.Popen("go build", shell=True)
        prc.wait()
        if prc.returncode != 0:
            print colored('ERROR: binary build failed', 'red')
            resp = False
        return resp


if __name__ == '__main__':
    GT = GoTests()
    if not GT.build():
        sys.exit(1)
    else:
        print "Build: ", colored("OK", "green")
    GT.run()
