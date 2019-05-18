package org.toughradius.common;

import java.util.concurrent.atomic.AtomicReference;

public class SpinLock {

    AtomicReference<Thread> owner = new AtomicReference<>();
    public void lock() {
        Thread cur = Thread.currentThread();
        while (!owner.compareAndSet(null, cur)){
        }
    }

    public void unLock() {
        Thread cur = Thread.currentThread();
        owner.compareAndSet(cur, null);
    }
}
