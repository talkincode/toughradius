package org.toughradius.config;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.apache.mina.core.filterchain.DefaultIoFilterChainBuilder;
import org.apache.mina.core.filterchain.IoFilter;
import org.apache.mina.filter.executor.ExecutorFilter;
import org.apache.mina.filter.logging.LoggingFilter;
import org.apache.mina.filter.logging.MdcInjectionFilter;
import org.apache.mina.transport.socket.DatagramSessionConfig;
import org.apache.mina.transport.socket.nio.NioDatagramAcceptor;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.toughradius.handler.RadiusAcctHandler;
import org.toughradius.handler.RadiusAuthHandler;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.util.LinkedHashMap;
import java.util.Map;

@Configuration
@ConfigurationProperties(prefix = "org.toughradius")
public class RadiusConfig {

    private Log logger = LogFactory.getLog(RadiusConfig.class);

    private String version;
    private int authport;
    private int acctport;
    private int trace;
    private int interimUpdate;
    private int maxSessionTimeout;
    private String ticketDir;
    private boolean running;
    private boolean isBillInput;
    private boolean isBillBackFlow;
    private boolean allowNegative;
    private int rejectdelay;
    private int ticketExpireDays;
    private int authPool;
    private int acctPool;
    private String statfile;

    /**
     * Radius 认证服务配置
     * @param radiusAuthHandler
     * @param authIoFilterChainBuilder
     * @return
     * @throws IOException
     */
    @Bean( destroyMethod = "unbind")
    public NioDatagramAcceptor nioAuthAcceptor(RadiusAuthHandler radiusAuthHandler, DefaultIoFilterChainBuilder authIoFilterChainBuilder) throws IOException {
        if(!running){
            logger.info("====== RadiusAuthServer not running =======");
            return null;
        }
        NioDatagramAcceptor nioAuthAcceptor = new NioDatagramAcceptor();
        nioAuthAcceptor.setDefaultLocalAddress(new InetSocketAddress(authport));
        DatagramSessionConfig dcfg = nioAuthAcceptor.getSessionConfig();
        dcfg.setReceiveBufferSize(8192);
        dcfg.setReadBufferSize(8192);
        dcfg.setSendBufferSize(4096);
        dcfg.setReuseAddress(true);
        nioAuthAcceptor.setFilterChainBuilder(authIoFilterChainBuilder);
        nioAuthAcceptor.setHandler(radiusAuthHandler);
        nioAuthAcceptor.bind();
        logger.info(String.format("====== RadiusAuthServer listen %s ======", authport));
        return nioAuthAcceptor;
    }


    @Bean
    public DefaultIoFilterChainBuilder authIoFilterChainBuilder(ExecutorFilter authExecutorFilter,
                                                                  MdcInjectionFilter authMdcInjectionFilter,
                                                                  LoggingFilter authLoggingFilter) {
        DefaultIoFilterChainBuilder authIoFilterChainBuilder = new DefaultIoFilterChainBuilder();
        Map<String, IoFilter> filters = new LinkedHashMap<>();
        filters.put("executor", authExecutorFilter);
//        filters.put("mdcInjectionFilter", authMdcInjectionFilter);
//        filters.put("loggingFilter", authLoggingFilter);
        authIoFilterChainBuilder.setFilters(filters);
        return authIoFilterChainBuilder;
    }


    @Bean
    public ExecutorFilter authExecutorFilter() {
        return new ExecutorFilter(getAuthPool());
    }

    @Bean
    public MdcInjectionFilter authMdcInjectionFilter() {
        return new MdcInjectionFilter(MdcInjectionFilter.MdcKey.remoteAddress);
    }

    @Bean
    public LoggingFilter authLoggingFilter() {
        return new LoggingFilter();
    }



    public int getAuthport() {
        return authport;
    }

    public void setAuthport(int authport) {
        this.authport = authport;
    }


    /**
     * Radius 记账服务配置
     * @param radiusAcctHandler
     * @param acctIoFilterChainBuilder
     * @return
     * @throws IOException
     */
    @Bean(destroyMethod = "unbind")
    public NioDatagramAcceptor nioAcctAcceptor(RadiusAcctHandler radiusAcctHandler, DefaultIoFilterChainBuilder acctIoFilterChainBuilder) throws IOException {
        if(!running){
            logger.info("====== RadiusAcctServer not running ======");
            return null;
        }
        NioDatagramAcceptor nioAcctAcceptor = new NioDatagramAcceptor();
        nioAcctAcceptor.setDefaultLocalAddress(new InetSocketAddress(acctport));
        DatagramSessionConfig dcfg = nioAcctAcceptor.getSessionConfig();
        dcfg.setReceiveBufferSize(8192);
        dcfg.setReadBufferSize(8192);
        dcfg.setSendBufferSize(4096);
        dcfg.setReuseAddress(true);
        nioAcctAcceptor.setFilterChainBuilder(acctIoFilterChainBuilder);
        nioAcctAcceptor.setHandler(radiusAcctHandler);
        nioAcctAcceptor.bind();
        logger.info(String.format("====== RadiusAcctServer listen %s ======", acctport));
        return nioAcctAcceptor;
    }


    @Bean
    public DefaultIoFilterChainBuilder acctIoFilterChainBuilder(ExecutorFilter acctExecutorFilter,
                                                                MdcInjectionFilter acctMdcInjectionFilter,
                                                                LoggingFilter acctLoggingFilter) {
        DefaultIoFilterChainBuilder acctIoFilterChainBuilder = new DefaultIoFilterChainBuilder();
        Map<String, IoFilter> filters = new LinkedHashMap<>();
        filters.put("executor", acctExecutorFilter);
//        filters.put("mdcInjectionFilter", acctMdcInjectionFilter);
//        filters.put("loggingFilter", acctLoggingFilter);
        acctIoFilterChainBuilder.setFilters(filters);
        return acctIoFilterChainBuilder;
    }

    @Bean
    public ExecutorFilter acctExecutorFilter() {
        return new ExecutorFilter(getAcctPool());
    }

    @Bean
    public MdcInjectionFilter acctMdcInjectionFilter() {
        return new MdcInjectionFilter(MdcInjectionFilter.MdcKey.remoteAddress);
    }


    @Bean
    public LoggingFilter acctLoggingFilter() {
        return new LoggingFilter();
    }

    public int getAcctport() {
        return acctport;
    }

    public void setAcctport(int acctport) {
        this.acctport = acctport;
    }


    public int getTrace() {
        return trace;
    }

    public void setTrace(int trace) {
        this.trace = trace;
    }

    public int getInterimUpdate() {
        return interimUpdate;
    }

    public void setInterimUpdate(int interimUpdate) {
        this.interimUpdate = interimUpdate;
    }

    public int getMaxSessionTimeout() {
        return maxSessionTimeout;
    }

    public void setMaxSessionTimeout(int maxSessionTimeout) {
        this.maxSessionTimeout = maxSessionTimeout;
    }

    public String getTicketDir() {
        return ticketDir;
    }

    public void setTicketDir(String ticketDir) {
        this.ticketDir = ticketDir;
    }

    public boolean isRunning() {
        return running;
    }

    public void setRunning(boolean running) {
        this.running = running;
    }

    public boolean isBillInput() {
        return isBillInput;
    }

    public void setBillInput(boolean billInput) {
        isBillInput = billInput;
    }

    public boolean isBillBackFlow() {
        return isBillBackFlow;
    }

    public void setBillBackFlow(boolean billBackFlow) {
        isBillBackFlow = billBackFlow;
    }

    public boolean isAllowNegative() {
        return allowNegative;
    }

    public void setAllowNegative(boolean allowNegative) {
        this.allowNegative = allowNegative;
    }

    public int getRejectdelay() {
        return rejectdelay;
    }

    public void setRejectdelay(int rejectdelay) {
        this.rejectdelay = rejectdelay;
    }

    public int getTicketExpireDays() {
        return ticketExpireDays;
    }

    public void setTicketExpireDays(int ticketExpireDays) {
        this.ticketExpireDays = ticketExpireDays;
    }

    public boolean isTraceEnabled(){
        return trace == 1;
    }

    public int getAuthPool() {
        return authPool;
    }

    public void setAuthPool(int authPool) {
        this.authPool = authPool;
    }

    public int getAcctPool() {
        return acctPool;
    }

    public void setAcctPool(int acctPool) {
        this.acctPool = acctPool;
    }

    public String getStatfile() {
        return statfile;
    }

    public void setStatfile(String statfile) {
        this.statfile = statfile;
    }

    public String getVersion() {
        return version;
    }

    public void setVersion(String version) {
        this.version = version;
    }
}
