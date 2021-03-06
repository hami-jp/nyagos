## Built-in commands

These commands have their alias. For example, `ls` => `__ls__`.

### `cd DRIVE:DIRECTORY`

Change the current working drive and directory.
No arguments, move to %HOME% or %USERPROFILE%.

* `cd -` : move the previous directory.
* `cd -N` (N:digit) : move the N-previous directory.
* `cd -h` , `cd ?` : listing directories stayed.
* `cd --history` : listing directories stayed all with no decoration

### `exit`

Quit NYAGOS.exe.

### `history [N]`

Display the history. No arguments, the last ten are displayed.

### `ln [-s] SRC DST`

Make hardlink or symbolic-link.
The alias 'lns' defined on `nyagos.d\lns.lua` shows UAC-dialog
and calls `ln -s`.

### `ls -OPTION FILES`

List the directory. Supported options are below:

* `-l` Long format
* `-F` Mark `/` after directories' name. `*' after executables' name.
* `-o` Enable color
* `-a` Print all files.
* `-R` Print Subdirectories recursively.
* `-1` Print filename only.
* `-t` Sort with last modified time.
* `-r` Revert sort order.
* `-h` With -l, print sizes in human readable format (e.g., 1K 234M 2G)
* `-S` Sort by file size

### `pwd`

Print the current woking drive and directory.

* `pwd -N` (N:digit) : print the N-previous directory.

### `set ENV=VAR`

Set the environment variable the value. When the value has any spaces,
you should `set "ENV=VAR"`.

* `PROMPT` ... The macro strings are compatible with CMD.EXE. Supported ANSI-ESCAPE SEQUENCE.

### `touch [-t [CC[YY]MMDDhhmm[.ss]]] [-r ref_file ] FILENAME(s)`

If FILENAME exists, update its timestamp, otherwise create it.

### `which [-a] COMMAND-NAME`

Report which file is executed.

* `-a` - report all executable on %PATH%

### `copy SOURCE-FILENAME DESTINATE-FILENAME`
### `copy SOURCE-FILENAME(S)... DESINATE-DIRECTORY`
### `move OLD-FILENAME NEW-FILENAME`
### `move SOURCE-FILENAME(S)... DESITINATE-DIRECTORY`
### `del FILE(S)...`
### `erase FILE(S)...`
### `mkdir [/p] NEWDIR(S)...`
### `rmdir [/s] DIR(S)...`
### `pushd [DIR]`
### `popd`
### `dirs`

These built-in commands are always asking with prompt when files are override or removed.

### `source BATCHFILENAME`

Execute the batch-file(*.cmd,*.bat) by CMD.exe and
import the environment variables and working directory
which CMD.exe changed.

We use . (one-period) as an alias of source.
