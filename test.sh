for dir in "nzb" "nntp" "watcher"; do
    pushd $dir
    go test
    popd
done
