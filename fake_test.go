package redigomock

import (
	"sort"
	"testing"

	"github.com/garyburd/redigo/redis"
)

func must(r interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return r
}

func assertInt(t *testing.T, result interface{}, expected int) {
	r, ok := result.(int)
	if !ok {
		t.Error("Invalid result type. Expect integer")
	}
	if r != expected {
		t.Errorf("Expect '%d'. Got '%d'", expected, r)
	}
}

func getError(r interface{}, err error) error {
	return err
}

func assertError(t *testing.T, err error) {
	if err == nil {
		t.Error("Error expected")
	}
}

func assertStrings(t *testing.T, result []string, expected []string, sorting bool) {
	r := make([]string, 0, len(result))
	e := make([]string, 0, len(expected))
	copy(r, result)
	copy(e, expected)
	if sorting {
		sort.Strings(r)
		sort.Strings(e)
	}
	if len(expected) != len(result) {
		t.Errorf("Excpected '%s', got '%s'", expected, result)
		return
	}
	for i := range e {
		if e[i] != r[i] {
			t.Errorf("Excpected '%s', got '%s'", expected, result)
			return
		}
	}
}

func TestFlushDb(t *testing.T) {
	c := NewFakeRedis()
	redis.String(c.Do("SET", "foo", "bar"))
	assertStrings(t, must(redis.Strings(c.Do("KEYS", "foo"))).([]string), []string{"foo"}, false)
	c.Do("FLUSHDB")
	assertStrings(t, must(redis.Strings(c.Do("KEYS", "foo"))).([]string), []string{}, false)
}

func TestSetThenGet(t *testing.T) {
	c := NewFakeRedis()
	redis.String(c.Do("SET", "foo", "bar"))
	r := must(redis.String(c.Do("GET", "foo"))).(string)
	if r != "bar" {
		t.Errorf("Expect 'bar'. Got '%s'", r)
	}
}

func TestMulti(t *testing.T) {
	c := NewFakeRedis()
	c.Send("MULTI")
	c.Send("SADD", "foo", "member1")
	c.Send("SADD", "foo", "member2")
	c.Send("SMEMBERS", "foo")
	res, err := c.Do("EXEC")
	assertStrings(t, must(redis.Strings(res.([]interface{})[3], err)).([]string), []string{"member1", "member2"}, true)
}

func TestGetThatNotExists(t *testing.T) {
	c := NewFakeRedis()
	r := must(c.Do("GET", "foo"))
	if r != nil {
		t.Errorf("Expect 'nil'. Got '%s'", r)
	}
}

func TestGetInvalidType(t *testing.T) {
	c := NewFakeRedis()
	assertInt(t, must(redis.Int(c.Do("SADD", "foo", "member1"))), 1)
	assertError(t, getError(c.Do("GET", "foo")))
}

func TestSadd(t *testing.T) {
	c := NewFakeRedis()
	assertInt(t, must(redis.Int(c.Do("SADD", "foo", "member1"))), 1)
	assertInt(t, must(redis.Int(c.Do("SADD", "foo", "member1"))), 0)
	assertStrings(t, must(redis.Strings(c.Do("SMEMBERS", "foo"))).([]string), []string{"member1"}, true)
	assertInt(t, must(redis.Int(c.Do("SADD", "foo", "member2", "member3"))), 2)
	assertStrings(t, must(redis.Strings(c.Do("SMEMBERS", "foo"))).([]string), []string{"member1", "member2", "member3"}, true)
	assertInt(t, must(redis.Int(c.Do("SADD", "foo", "member3", "member4"))), 1)
	assertStrings(t, must(redis.Strings(c.Do("SMEMBERS", "foo"))).([]string), []string{"member1", "member2", "member3", "member4"}, true)
}

func TestSrem(t *testing.T) {
	c := NewFakeRedis()
	c.Do("SADD", "foo", "member1", "member2", "member3", "member4")
	assertStrings(t, must(redis.Strings(c.Do("SMEMBERS", "foo"))).([]string), []string{"member1", "member2", "member3", "member4"}, true)
	assertInt(t, must(redis.Int(c.Do("SREM", "foo", "member1"))), 1)
	assertStrings(t, must(redis.Strings(c.Do("SMEMBERS", "foo"))).([]string), []string{"member2", "member3", "member4"}, true)
	assertInt(t, must(redis.Int(c.Do("SREM", "foo", "member1"))), 0)
	assertInt(t, must(redis.Int(c.Do("SREM", "foo", "member2", "member3"))), 2)
	assertStrings(t, must(redis.Strings(c.Do("SMEMBERS", "foo"))).([]string), []string{"member4"}, true)
	assertInt(t, must(redis.Int(c.Do("SREM", "foo", "member3", "member4"))), 1)
	assertStrings(t, must(redis.Strings(c.Do("SMEMBERS", "foo"))).([]string), []string{}, false)
	assertInt(t, must(redis.Int(c.Do("SREM", "foo", "member3", "member4"))), 0)
}

func TestSUnion(t *testing.T) {
	c := NewFakeRedis()
	c.Do("SADD", "foo", "member1", "member2")
	c.Do("SADD", "bar", "member2", "member3")
	assertStrings(t, must(redis.Strings(c.Do("SUNION", "foo", "bar"))).([]string), []string{"member1", "member2", "member3"}, true)
}

func TestZadd(t *testing.T) {
	c := NewFakeRedis()
	c.Do("ZADD", "foo", 4, "four")
	c.Do("ZADD", "foo", 3, "three")
	assertInt(t, must(redis.Int(c.Do("ZADD", "foo", 2, "two", 1, "one", 0, "zero"))), 3)
	assertStrings(t, must(redis.Strings(c.Do("ZRANGE", "foo", 0, -1))).([]string), []string{"zero", "one", "two", "three", "four"}, false)
	assertInt(t, must(redis.Int(c.Do("ZADD", "foo", 7, "zero", 1, "one", 5, "five"))), 1)
	assertStrings(t, must(redis.Strings(c.Do("ZRANGE", "foo", 0, -1))).([]string), []string{"one", "two", "three", "four", "five", "zero"}, false)
}

func TestZcount(t *testing.T) {
	c := NewFakeRedis()
	c.Do("ZADD", "foo", 1, "one")
	c.Do("ZADD", "foo", 2, "three")
	c.Do("ZADD", "foo", 5, "five")
	assertInt(t, must(redis.Int(c.Do("ZCOUNT", "foo", 2, 4))), 1)
	assertInt(t, must(redis.Int(c.Do("ZCOUNT", "foo", 1, 4))), 2)
	assertInt(t, must(redis.Int(c.Do("ZCOUNT", "foo", 0, 5))), 3)
	assertInt(t, must(redis.Int(c.Do("ZCOUNT", "foo", 4, "+inf"))), 1)
	assertInt(t, must(redis.Int(c.Do("ZCOUNT", "foo", "-inf", 4))), 2)
	assertInt(t, must(redis.Int(c.Do("ZCOUNT", "foo", "-inf", "+inf"))), 3)
}

func TestZCountExclusive(t *testing.T) {
	c := NewFakeRedis()
	c.Do("ZADD", "foo", 1, "one")
	c.Do("ZADD", "foo", 2, "three")
	c.Do("ZADD", "foo", 5, "five")
	assertInt(t, must(redis.Int(c.Do("ZCOUNT", "foo", "-inf", "(2"))), 1)
	assertInt(t, must(redis.Int(c.Do("ZCOUNT", "foo", "-inf", 2))), 2)
	assertInt(t, must(redis.Int(c.Do("ZCOUNT", "foo", "(5", "+inf"))), 0)
	assertInt(t, must(redis.Int(c.Do("ZCOUNT", "foo", "(1", 5))), 2)
	assertInt(t, must(redis.Int(c.Do("ZCOUNT", "foo", "(2", "(5"))), 0)
	assertInt(t, must(redis.Int(c.Do("ZCOUNT", "foo", "(1", "(5"))), 1)
	assertInt(t, must(redis.Int(c.Do("ZCOUNT", "foo", 2, "(5"))), 1)
}

// TODO: test_getbit(self):
// TODO: test_multiple_bits_set(self):
// TODO: test_unset_bits(self):
// TODO: test_setbits_and_getkeys(self):
// TODO: test_bitcount(self):
// TODO: test_getset_not_exist(self):
// TODO: test_getset_exists(self):
// TODO: test_setitem_getitem(self):
// TODO: test_strlen(self):
// TODO: test_substr(self):
// TODO: test_substr_noexist_key(self):
// TODO: test_append(self):
// TODO: test_append_with_no_preexisting_key(self):
// TODO: test_incr_with_no_preexisting_key(self):
// TODO: test_incr_by(self):
// TODO: test_incr_preexisting_key(self):
// TODO: test_incr_bad_type(self):
// TODO: test_incrbyfloat(self):
// TODO: test_incrbyfloat_with_noexist(self):
// TODO: test_incrbyfloat_bad_type(self):
// TODO: test_decr(self):
// TODO: test_decr_newkey(self):
// TODO: test_decr_badtype(self):
// TODO: test_exists(self):
// TODO: test_contains(self):
// TODO: test_rename(self):
// TODO: test_rename_nonexistent_key(self):
// TODO: test_renamenx_doesnt_exist(self):
// TODO: test_rename_does_exist(self):
// TODO: test_mget(self):
// TODO: test_mgset_with_no_keys_raises_error(self):
// TODO: test_mset_with_no_keys_raises_error(self):
// TODO: test_mset(self):
// TODO: test_mset_accepts_kwargs(self):
// TODO: test_msetnx(self):
// TODO: test_setex(self):
// TODO: test_setex_using_timedelta(self):
// TODO: test_setnx(self):
// TODO: test_delete(self):
// TODO: test_echo(self):
// TODO: test_delete_expire(self):
// TODO: test_delete_multiple(self):
// TODO: test_delete_nonexistent_key(self):
// TODO: test_rpush_then_lrange_with_nested_list1(self):
// TODO: test_rpush_then_lrange_with_nested_list2(self):
// TODO: test_rpush_then_lrange_with_nested_list3(self):
// TODO: test_lpush_then_lrange_all(self):
// TODO: test_lpush_then_lrange_portion(self):
// TODO: test_lpush_key_does_not_exist(self):
// TODO: test_lpush_with_nonstr_key(self):
// TODO: test_llen(self):
// TODO: test_llen_no_exist(self):
// TODO: test_lrem_postitive_count(self):
// TODO: test_lrem_negative_count(self):
// TODO: test_lrem_zero_count(self):
// TODO: test_lrem_default_value(self):
// TODO: test_lrem_does_not_exist(self):
// TODO: test_lrem_return_value(self):
// TODO: test_rpush(self):
// TODO: test_lpop(self):
// TODO: test_lpop_empty_list(self):
// TODO: test_lset(self):
// TODO: test_lset_index_out_of_range(self):
// TODO: test_rpushx(self):
// TODO: test_ltrim(self):
// TODO: test_ltrim_with_non_existent_key(self):
// TODO: test_lindex(self):
// TODO: test_lpushx(self):
// TODO: test_rpop(self):
// TODO: test_linsert(self):
// TODO: test_rpoplpush(self):
// TODO: test_rpoplpush_to_nonexistent_destination(self):
// TODO: test_blpop_single_list(self):
// TODO: test_blpop_test_multiple_lists(self):
// TODO: test_blpop_allow_single_key(self):
// TODO: test_brpop_test_multiple_lists(self):
// TODO: test_brpop_single_key(self):
// TODO: test_brpoplpush_multi_keys(self):
// TODO: test_blocking_operations_when_empty(self):
// TODO: test_hset_then_hget(self):
// TODO: test_hset_update(self):
// TODO: test_hgetall(self):
// TODO: test_hgetall_with_tuples(self):
// TODO: test_hgetall_empty_key(self):
// TODO: test_hexists(self):
// TODO: test_hkeys(self):
// TODO: test_hlen(self):
// TODO: test_hvals(self):
// TODO: test_hmget(self):
// TODO: test_hdel(self):
// TODO: test_hincrby(self):
// TODO: test_hincrby_with_no_starting_value(self):
// TODO: test_hincrby_with_range_param(self):
// TODO: test_hincrbyfloat(self):
// TODO: test_hincrbyfloat_with_no_starting_value(self):
// TODO: test_hincrbyfloat_with_range_param(self):
// TODO: test_hincrbyfloat_on_non_float_value_raises_error(self):
// TODO: test_hincrbyfloat_with_non_float_amount_raises_error(self):
// TODO: test_hsetnx(self):
// TODO: test_hmsetset_empty_raises_error(self):
// TODO: test_hmsetset(self):
// TODO: test_hmset_convert_values(self):
// TODO: test_hmset_does_not_mutate_input_params(self):
// TODO: test_scan_single(self):
// TODO: test_scan_iter_single_page(self):
// TODO: test_scan_iter_multiple_pages(self):
// TODO: test_scan_iter_multiple_pages_with_match(self):
// TODO: test_scan_multiple_pages_with_count_arg(self):
// TODO: test_scan_all_in_single_call(self):
// TODO: test_scard(self):
// TODO: test_sdiff(self):
// TODO: test_sdiff_one_key(self):
// TODO: test_sdiff_empty(self):
// TODO: test_sdiffstore(self):
// TODO: test_sinter(self):
// TODO: test_sinterstore(self):
// TODO: test_sismember(self):
// TODO: test_smembers(self):
// TODO: test_smove(self):
// TODO: test_smove_non_existent_key(self):
// TODO: test_spop(self):
// TODO: test_srandmember(self):
// TODO: test_sunionstore(self):
// TODO: test_zadd(self):
// TODO: test_zadd_uses_str(self):
// TODO: test_zadd_errors(self):
// TODO: test_zadd_multiple(self):
// TODO: test_zrange_same_score(self):
// TODO: test_zcard(self):
// TODO: test_zcard_non_existent_key(self):
// TODO: test_zincrby(self):
// TODO: test_zrange_descending(self):
// TODO: test_zrange_descending_with_scores(self):
// TODO: test_zrange_with_positive_indices(self):
// TODO: test_zrank(self):
// TODO: test_zrank_non_existent_member(self):
// TODO: test_zrem(self):
// TODO: test_zrem_non_existent_member(self):
// TODO: test_zrem_numeric_member(self):
// TODO: test_zscore(self):
// TODO: test_zscore_non_existent_member(self):
// TODO: test_zrevrank(self):
// TODO: test_zrevrank_non_existent_member(self):
// TODO: test_zrevrange(self):
// TODO: test_zrevrange_sorted_keys(self):
// TODO: test_zrangebyscore(self):
// TODO: test_zrangebysore_exclusive(self):
// TODO: test_zrangebyscore_raises_error(self):
// TODO: test_zrangebyscore_slice(self):
// TODO: test_zrangebyscore_withscores(self):
// TODO: test_zrevrangebyscore(self):
// TODO: test_zrevrangebyscore_exclusive(self):
// TODO: test_zrevrangebyscore_raises_error(self):
// TODO: test_zremrangebyrank(self):
// TODO: test_zremrangebyrank_negative_indices(self):
// TODO: test_zremrangebyrank_out_of_bounds(self):
// TODO: test_zremrangebyscore(self):
// TODO: test_zremrangebyscore_exclusive(self):
// TODO: test_zremrangebyscore_raises_error(self):
// TODO: test_zremrangebyscore_badkey(self):
// TODO: test_zunionstore(self):
// TODO: test_zunionstore_sum(self):
// TODO: test_zunionstore_max(self):
// TODO: test_zunionstore_min(self):
// TODO: test_zunionstore_weights(self):
// TODO: test_zunionstore_mixed_set_types(self):
// TODO: test_zunionstore_badkey(self):
// TODO: test_zinterstore(self):
// TODO: test_zinterstore_mixed_set_types(self):
// TODO: test_zinterstore_max(self):
// TODO: test_zinterstore_onekey(self):
// TODO: test_zinterstore_nokey(self):
// TODO: test_zunionstore_nokey(self):
// TODO: test_multidb(self):
// TODO: test_basic_sort(self):
// TODO: test_empty_sort(self):
// TODO: test_sort_range_offset_range(self):
// TODO: test_sort_range_offset_range_and_desc(self):
// TODO: test_sort_range_offset_norange(self):
// TODO: test_sort_range_with_large_range(self):
// TODO: test_sort_descending(self):
// TODO: test_sort_alpha(self):
// TODO: test_foo(self):
// TODO: test_sort_with_store_option(self):
// TODO: test_sort_with_by_and_get_option(self):
// TODO: test_sort_with_hash(self):
// TODO: test_sort_with_set(self):
// TODO: test_pipeline(self):
// TODO: test_pipeline_ignore_errors(self):
// TODO: test_multiple_successful_watch_calls(self):
// TODO: test_pipeline_non_transactional(self):
// TODO: test_pipeline_raises_when_watched_key_changed(self):
// TODO: test_pipeline_succeeds_despite_unwatched_key_changed(self):
// TODO: test_pipeline_succeeds_when_watching_nonexistent_key(self):
// TODO: test_watch_state_is_cleared_across_multiple_watches(self):
// TODO: test_pipeline_proxies_to_redis_object(self):
// TODO: test_pipeline_as_context_manager(self):
// TODO: test_pipeline_transaction_shortcut(self):
// TODO: test_key_patterns(self):
// TODO: test_ping(self):
// TODO: test_type(self):
// TODO: test_pubsub_subscribe(self):
// TODO: test_pubsub_psubscribe(self):
// TODO: test_pubsub_unsubscribe(self):
// TODO: test_pubsub_punsubscribe(self):
// TODO: test_pubsub_listen(self):
// TODO: test_setex(self):
// TODO: test_setex_using_timedelta(self):
// TODO: test_lrem_postitive_count(self):
// TODO: test_lrem_negative_count(self):
// TODO: test_lrem_zero_count(self):
// TODO: test_lrem_default_value(self):
// TODO: test_lrem_does_not_exist(self):
// TODO: test_lrem_return_value(self):
// TODO: test_zadd_deprecated(self):
// TODO: test_zadd_missing_required_params(self):
// TODO: test_zadd_with_single_keypair(self):
// TODO: test_zadd_with_multiple_keypairs(self):
// TODO: test_zadd_with_name_is_non_string(self):
// TODO: test_set_nx_doesnt_set_value_twice(self):
// TODO: test_set_xx_set_value_when_exists(self):
// TODO: test_set_ex_should_expire_value(self):
// TODO: test_set_px_should_expire_value(self):
// TODO: test_psetex_expire_value(self):
// TODO: test_psetex_expire_value_using_timedelta(self):
// TODO: test_expire_should_expire_key(self):
// TODO: test_expire_should_return_true_for_existing_key(self):
// TODO: test_expire_should_return_false_for_missing_key(self):
// TODO: test_expire_should_expire_key_using_timedelta(self):
// TODO: test_expireat_should_expire_key_by_datetime(self):
// TODO: test_expireat_should_expire_key_by_timestamp(self):
// TODO: test_expireat_should_return_true_for_existing_key(self):
// TODO: test_expireat_should_return_false_for_missing_key(self):
// TODO: test_ttl_should_return_none_for_non_expiring_key(self):
// TODO: test_ttl_should_return_value_for_expiring_key(self):
// TODO: test_pttl_should_return_none_for_non_expiring_key(self):
// TODO: test_pttl_should_return_value_for_expiring_key(self):
