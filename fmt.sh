for dir in "nzb" "nntp" "watcher"; do
    pushd $dir
    go fmt
    popd
done
