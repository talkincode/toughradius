package org.toughradius.config;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.apache.mina.core.filterchain.DefaultIoFilterChainBuilder;
import org.apache.mina.core.filterchain.IoFilter;
import org.apache.mina.core.session.IoSession;
import org.apache.mina.filter.executor.ExecutorFilter;
import org.apache.mina.filter.ssl.KeyStoreFactory;
import org.apache.mina.filter.ssl.SslContextFactory;
import org.apache.mina.filter.ssl.SslFilter;
import org.apache.mina.transport.socket.SocketSessionConfig;
import org.apache.mina.transport.socket.nio.NioSocketAcceptor;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.toughradius.common.DefaultThreadFactory;
import org.toughradius.handler.RadsecHandler;

import javax.net.ssl.SSLContext;
import java.io.File;
import java.io.IOException;
import java.net.InetSocketAddress;
import java.security.KeyStore;
import java.util.*;
import java.util.concurrent.TimeUnit;

@Configuration
@ConfigurationProperties(prefix = "org.toughradius.radsec")
public class RadsecConfig {

    private Log logger = LogFactory.getLog(RadsecConfig.class);

    private boolean enabled;
    private int port;
    private int pool;

    private String keyStoreFile;

    /**
     * Radius 认证服务配置
     * @param radsecHandler
     * @return
     * @throws IOException
     */
    @Bean( destroyMethod = "unbind")
    public NioSocketAcceptor nioRadsecAcceptor(RadsecHandler radsecHandler) throws IOException {
        if(!enabled){
            logger.info("====== RadsecServer not running =======");
            return null;
        }
        NioSocketAcceptor nioRadsecAcceptor = new NioSocketAcceptor();
        nioRadsecAcceptor.setDefaultLocalAddress(new InetSocketAddress(port));
        SocketSessionConfig dcfg = nioRadsecAcceptor.getSessionConfig();
        dcfg.setReceiveBufferSize(33554432);
        dcfg.setReadBufferSize(8192);
        dcfg.setSendBufferSize(8192);
        dcfg.setBothIdleTime(0);
        dcfg.setReuseAddress(true);

        DefaultIoFilterChainBuilder authIoFilterChainBuilder = new DefaultIoFilterChainBuilder();
        ExecutorFilter authExecutorFilter = new ExecutorFilter(8, getPool(), 60, TimeUnit.SECONDS,new DefaultThreadFactory("radsecExecutorFilter",Thread.MAX_PRIORITY));
        Map<String, IoFilter> filters = new LinkedHashMap<>();
        filters.put("executor", authExecutorFilter);
        SslFilter sslFilter = new SslFilter(new SSLContextGenerator().getSslContext());
        filters.put("sslFilter", sslFilter);
        authIoFilterChainBuilder.setFilters(filters);
        nioRadsecAcceptor.setFilterChainBuilder(authIoFilterChainBuilder);
        nioRadsecAcceptor.setHandler(radsecHandler);
        nioRadsecAcceptor.bind();
        logger.info(String.format("====== RadsecServer listen %s ======", port));
        return nioRadsecAcceptor;
    }

    public boolean isEnabled() {
        return enabled;
    }

    public void setEnabled(boolean enabled) {
        this.enabled = enabled;
    }

    public int getPort() {
        return port;
    }

    public void setPort(int port) {
        this.port = port;
    }

    public int getPool() {
        return pool;
    }

    public void setPool(int pool) {
        this.pool = pool;
    }

    public String getKeyStoreFile() {
        return keyStoreFile;
    }

    public void setKeyStoreFile(String keyStoreFile) {
        this.keyStoreFile = keyStoreFile;
    }

    /**
     * @author behindjava.com
     */
    class SSLContextGenerator
    {
        public SSLContext getSslContext()
        {
            SSLContext sslContext = null;
            try
            {
                File _keyStoreFile = new File(getKeyStoreFile());
                if(!_keyStoreFile.exists()){
                    _keyStoreFile = new File(Objects.requireNonNull(SSLContextGenerator.class.getClassLoader().getResource("radsec/server.p12")).toURI());
                }

                logger.info("Use keyStoreFile: " + _keyStoreFile.getAbsolutePath());
                final KeyStoreFactory keyStoreFactory = new KeyStoreFactory();
                keyStoreFactory.setDataFile(_keyStoreFile);
//                keyStoreFactory.setType("PKCS12");
                keyStoreFactory.setPassword("radsec");

                final KeyStoreFactory trustStoreFactory = new KeyStoreFactory();
                trustStoreFactory.setDataFile(_keyStoreFile);
//                trustStoreFactory.setType("PKCS12");
                trustStoreFactory.setPassword("radsec");

                final SslContextFactory sslContextFactory = new SslContextFactory();
                final KeyStore keyStore = keyStoreFactory.newInstance();
                sslContextFactory.setKeyManagerFactoryKeyStore(keyStore);

                final KeyStore trustStore = trustStoreFactory.newInstance();
                sslContextFactory.setTrustManagerFactoryKeyStore(trustStore);
                sslContextFactory.setKeyManagerFactoryKeyStorePassword("radsec");
                sslContext = sslContextFactory.newInstance();
                logger.info("SSL provider is: " + sslContext.getProvider());
            }
            catch (Exception ex)
            {
                logger.error("getSslContext error",ex);
            }
            return sslContext;
        }
    }
}
