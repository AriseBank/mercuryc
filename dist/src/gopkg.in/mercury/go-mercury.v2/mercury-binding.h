// Copyright © 2013, 2014, The Go-MERCURY Authors. All rights reserved.
// Use of this source code is governed by a LGPLv2.1
// license that can be found in the LICENSE file.

extern bool go_mercury_add_device_node(struct mercury_container *c, const char *src_path, const char *dest_path);
extern void go_mercury_clear_config(struct mercury_container *c);
extern bool go_mercury_clear_config_item(struct mercury_container *c, const char *key);
extern bool go_mercury_clone(struct mercury_container *c, const char *newname, const char *mercurypath, int flags, const char *bdevtype);
extern bool go_mercury_console(struct mercury_container *c, int ttynum, int stdinfd, int stdoutfd, int stderrfd, int escape);
extern bool go_mercury_create(struct mercury_container *c, const char *t, const char *bdevtype, int flags, char * const argv[]);
extern bool go_mercury_defined(struct mercury_container *c);
extern bool go_mercury_destroy(struct mercury_container *c);
extern bool go_mercury_destroy_with_snapshots(struct mercury_container *c);
extern bool go_mercury_freeze(struct mercury_container *c);
extern bool go_mercury_load_config(struct mercury_container *c, const char *alt_file);
extern bool go_mercury_may_control(struct mercury_container *c);
extern bool go_mercury_reboot(struct mercury_container *c);
extern bool go_mercury_remove_device_node(struct mercury_container *c, const char *src_path, const char *dest_path);
extern bool go_mercury_rename(struct mercury_container *c, const char *newname);
extern bool go_mercury_running(struct mercury_container *c);
extern bool go_mercury_save_config(struct mercury_container *c, const char *alt_file);
extern bool go_mercury_set_cgroup_item(struct mercury_container *c, const char *key, const char *value);
extern bool go_mercury_set_config_item(struct mercury_container *c, const char *key, const char *value);
extern bool go_mercury_set_config_path(struct mercury_container *c, const char *path);
extern bool go_mercury_shutdown(struct mercury_container *c, int timeout);
extern bool go_mercury_snapshot_destroy(struct mercury_container *c, const char *snapname);
extern bool go_mercury_snapshot_destroy_all(struct mercury_container *c);
extern bool go_mercury_snapshot_restore(struct mercury_container *c, const char *snapname, const char *newname);
extern bool go_mercury_start(struct mercury_container *c, int useinit, char * const argv[]);
extern bool go_mercury_stop(struct mercury_container *c);
extern bool go_mercury_unfreeze(struct mercury_container *c);
extern bool go_mercury_wait(struct mercury_container *c, const char *state, int timeout);
extern bool go_mercury_want_close_all_fds(struct mercury_container *c, bool state);
extern bool go_mercury_want_daemonize(struct mercury_container *c, bool state);
extern char* go_mercury_config_file_name(struct mercury_container *c);
extern char* go_mercury_get_cgroup_item(struct mercury_container *c, const char *key);
extern char* go_mercury_get_config_item(struct mercury_container *c, const char *key);
extern char** go_mercury_get_interfaces(struct mercury_container *c);
extern char** go_mercury_get_ips(struct mercury_container *c, const char *interface, const char *family, int scope);
extern char* go_mercury_get_keys(struct mercury_container *c, const char *key);
extern char* go_mercury_get_running_config_item(struct mercury_container *c, const char *key);
extern const char* go_mercury_get_config_path(struct mercury_container *c);
extern const char* go_mercury_state(struct mercury_container *c);
extern int go_mercury_attach_run_wait(struct mercury_container *c,
		bool clear_env,
		int namespaces,
		long personality,
		uid_t uid, gid_t gid,
		int stdinfd, int stdoutfd, int stderrfd,
		char *initial_cwd,
		char **extra_env_vars,
		char **extra_keep_env,
		const char * const argv[]);
extern int go_mercury_attach(struct mercury_container *c,
		bool clear_env,
		int namespaces,
		long personality,
		uid_t uid, gid_t gid,
		int stdinfd, int stdoutfd, int stderrfd,
		char *initial_cwd,
		char **extra_env_vars,
		char **extra_keep_env);
extern int go_mercury_attach_no_wait(struct mercury_container *c,
		bool clear_env,
		int namespaces,
		long personality,
		uid_t uid, gid_t gid,
		int stdinfd, int stdoutfd, int stderrfd,
		char *initial_cwd,
		char **extra_env_vars,
		char **extra_keep_env,
		const char * const argv[],
		pid_t *attached_pid);
extern int go_mercury_console_getfd(struct mercury_container *c, int ttynum);
extern int go_mercury_snapshot_list(struct mercury_container *c, struct mercury_snapshot **ret);
extern int go_mercury_snapshot(struct mercury_container *c);
extern pid_t go_mercury_init_pid(struct mercury_container *c);
extern bool go_mercury_checkpoint(struct mercury_container *c, char *directory, bool stop, bool verbose);
extern bool go_mercury_restore(struct mercury_container *c, char *directory, bool verbose);
extern bool go_mercury_config_item_is_supported(const char *key);

/* n.b. that we're just adding the fields here to shorten the definition
 * of go_mercury_migrate; in the case where we don't have the ->migrate API call,
 * we don't want to have to pass all the arguments in to let conditional
 * compilation handle things, but the call will still fail
 */
#if MERCURY_VERSION_MAJOR != 2
struct migrate_opts {
	char *directory;
	bool verbose;
	bool stop;
	char *predump_dir;
};
#endif

/* This is a struct that we can add "extra" (i.e. options added after 2.0.0)
 * migrate options to, so that we don't have to have a massive function
 * signature when the list of options grows.
 */
struct extra_migrate_opts {
	bool preserves_inodes;
	char *action_script;
	uint64_t ghost_limit;
};
int go_mercury_migrate(struct mercury_container *c, unsigned int cmd, struct migrate_opts *opts, struct extra_migrate_opts *extras);

extern bool go_mercury_attach_interface(struct mercury_container *c, const char *dev, const char *dst_dev);
extern bool go_mercury_detach_interface(struct mercury_container *c, const char *dev, const char *dst_dev);
