#!/usr/bin/python
#coding:utf-8
import test_config


def test_user_trace():
    from radiusd.admin import UserTrace
    trace = UserTrace()
    assert trace
    assert trace.size_info() == (0,0)
    trace.push("test", {})
    assert trace.size_info() == (1,1)
    for i in range(30):
        trace.push("test%s"%i, {})
    assert trace.size_info() == (31,20),trace.size_info()
    for uc in trace.user_cache:
        assert len(trace.user_cache[uc]) == 1
        break

    trace2 = UserTrace()
    pkt = {'id':1}
    trace2.push("test", pkt)
    assert trace2.get_global_msg() == pkt
    assert trace2.get_user_msg('') == []
    assert trace2.get_user_msg('test') == [pkt]


if __name__ == '__main__':
    test_user_trace()