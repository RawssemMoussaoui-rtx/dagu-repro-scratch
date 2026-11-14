/*
 * Shared DTrace helper for FD state tracking in one target process.
 * Run with: dtrace -q -s dtrace/trace.d -p <pid>
 */
#pragma D option quiet

BEGIN
{
    /* Descriptors 0-2 are assumed open when tracing starts. */
    fd_valid[0] = 1;
    fd_valid[1] = 1;
    fd_valid[2] = 1;
    fd_gen[0] = 1;
    fd_gen[1] = 2;
    fd_gen[2] = 3;
    gen_seq = 3;
}

syscall::open:entry
/pid == $target && $target > 0/
{
    self->open_path = copyinstr(arg0);
}

syscall::open:return
/pid == $target && $target > 0 && arg0 >= 0/
{
    this->fd = (int)arg0;
    fd_valid[this->fd] = 1;
    gen_seq = gen_seq + 1;
    fd_gen[this->fd] = gen_seq;
    printf("target=%d event=open fd=%d gen=%d valid=1 path=%s\n",
        $target, this->fd, fd_gen[this->fd], self->open_path);
}

syscall::openat:entry
/pid == $target && $target > 0/
{
    self->openat_dirfd = arg0;
    self->openat_path = copyinstr(arg1);
}

syscall::openat:return
/pid == $target && $target > 0 && arg0 >= 0/
{
    this->fd = (int)arg0;
    fd_valid[this->fd] = 1;
    gen_seq = gen_seq + 1;
    fd_gen[this->fd] = gen_seq;
    printf("target=%d event=openat dirfd=%d fd=%d gen=%d valid=1 path=%s\n",
        $target, self->openat_dirfd, this->fd, fd_gen[this->fd], self->openat_path);
}

syscall::close:entry
/pid == $target && $target > 0/
{
    self->close_fd = arg0;
}

syscall::close:return
/pid == $target && $target > 0 && arg0 == 0/
{
    fd_valid[self->close_fd] = 0;
    printf("target=%d event=close fd=%d gen=%d valid=0\n",
        $target, self->close_fd, fd_gen[self->close_fd]);
}

syscall::dup:entry
/pid == $target && $target > 0/
{
    self->dup_oldfd = arg0;
}

syscall::dup:return
/pid == $target && $target > 0 && arg0 >= 0/
{
    this->newfd = (int)arg0;
    fd_valid[this->newfd] = 1;
    gen_seq = gen_seq + 1;
    fd_gen[this->newfd] = gen_seq;
    printf("target=%d event=dup oldfd=%d newfd=%d gen=%d valid=1\n",
        $target, self->dup_oldfd, this->newfd, fd_gen[this->newfd]);
}

syscall::dup2:entry
/pid == $target && $target > 0/
{
    self->dup2_oldfd = arg0;
    self->dup2_newfd = arg1;
}

syscall::dup2:return
/pid == $target && $target > 0 && arg0 >= 0/
{
    fd_valid[self->dup2_newfd] = fd_valid[self->dup2_oldfd];
    gen_seq = gen_seq + 1;
    fd_gen[self->dup2_newfd] = gen_seq;
    printf("target=%d event=dup2 oldfd=%d newfd=%d gen=%d valid=%d\n",
        $target, self->dup2_oldfd, self->dup2_newfd,
        fd_gen[self->dup2_newfd], fd_valid[self->dup2_newfd]);
}

syscall::fcntl:entry
/pid == $target && $target > 0/
{
    self->fcntl_fd = arg0;
    self->fcntl_cmd = arg1;
    self->fcntl_arg = arg2;
}

syscall::fcntl:return
/pid == $target && $target > 0 && arg0 >= 0 &&
    (self->fcntl_cmd == 0 || self->fcntl_cmd == 1030)/
{
    this->newfd = (int)arg0;
    fd_valid[this->newfd] = 1;
    gen_seq = gen_seq + 1;
    fd_gen[this->newfd] = gen_seq;
    printf("target=%d event=fcntl cmd=%d fd=%d newfd=%d gen=%d valid=1\n",
        $target, self->fcntl_cmd, self->fcntl_fd, this->newfd, fd_gen[this->newfd]);
}
