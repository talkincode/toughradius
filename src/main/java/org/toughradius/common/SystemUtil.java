package org.toughradius.common;

import com.sun.management.OperatingSystemMXBean;

import java.io.*;
import java.lang.management.ManagementFactory;
import java.math.BigDecimal;
import java.util.*;

public class SystemUtil {

    private static String osName = System.getProperty("os.name");
    private static final int CPUTIME = 500;
    private static final int PERCENT = 100;
    private static final int FAULTLENGTH = 10;

    public static long getUptime(){
        return ManagementFactory.getRuntimeMXBean().getUptime();
    }

    /**
     * 功能：获取Linux和Window系统cpu使用率
     * */
    public static double getCpuUsage() {
// 如果是window系统
        if (osName.toLowerCase().contains("windows")
                || osName.toLowerCase().contains("win")) {
            return 30.0;
//            try {
//                String procCmd = System.getenv("windir")
//                        + "//system32//wbem//wmic.exe process get Caption,CommandLine,KernelModeTime,ReadOperationCount,ThreadCount,UserModeTime,WriteOperationCount";
//// 取进程信息
//                long[] c0 = readCpu(Runtime.getRuntime().exec(procCmd));//第一次读取CPU信息
//                Thread.sleep(CPUTIME);//睡500ms
//                long[] c1 = readCpu(Runtime.getRuntime().exec(procCmd));//第二次读取CPU信息
//                if (c0 != null && c1 != null) {
//                    long idletime = c1[0] - c0[0];//空闲时间
//                    long busytime = c1[1] - c0[1];//使用时间
//                    Double cpusage = Double.valueOf(PERCENT * (busytime) * 1.0 / (busytime + idletime));
//                    BigDecimal b1 = new BigDecimal(cpusage);
//                    double cpuUsage = b1.setScale(2, BigDecimal.ROUND_HALF_UP).doubleValue();
//                    return cpuUsage;
//                } else {
//                    return 0.0;
//                }
//            } catch (Exception ex) {
//                logger.error("",ex);
//                return 0.0;
//            }

        } else {

            try {
                Map<?, ?> map1 = SystemUtil.cpuinfo();
                Thread.sleep(CPUTIME);
                Map<?, ?> map2 = SystemUtil.cpuinfo();

                long user1 = Long.parseLong(map1.get("user").toString());
                long nice1 = Long.parseLong(map1.get("nice").toString());
                long system1 = Long.parseLong(map1.get("system").toString());
                long idle1 = Long.parseLong(map1.get("idle").toString());

                long user2 = Long.parseLong(map2.get("user").toString());
                long nice2 = Long.parseLong(map2.get("nice").toString());
                long system2 = Long.parseLong(map2.get("system").toString());
                long idle2 = Long.parseLong(map2.get("idle").toString());

                long total1 = user1 + system1 + nice1;
                long total2 = user2 + system2 + nice2;
                float total = total2 - total1;

                long totalIdle1 = user1 + nice1 + system1 + idle1;
                long totalIdle2 = user2 + nice2 + system2 + idle2;
                float totalidle = totalIdle2 - totalIdle1;

                float cpusage = (total / totalidle) * 100;

                BigDecimal b1 = new BigDecimal(cpusage);
                double cpuUsage = b1.setScale(2, BigDecimal.ROUND_HALF_UP).doubleValue();
                return cpuUsage;
            } catch (InterruptedException e) {
                e.printStackTrace();
            }
        }
        return 0;
    }

    /**
     * 功能：Linux CPU使用信息
     * */
    public static Map<?, ?> cpuinfo() {
        InputStreamReader inputs = null;
        BufferedReader buffer = null;
        Map<String, Object> map = new HashMap<String, Object>();
        try {
            inputs = new InputStreamReader(new FileInputStream("/proc/stat"));
            buffer = new BufferedReader(inputs);
            String line = "";
            while (true) {
                line = buffer.readLine();
                if (line == null) {
                    break;
                }
                if (line.startsWith("cpu")) {
                    StringTokenizer tokenizer = new StringTokenizer(line);
                    List<String> temp = new ArrayList<String>();
                    while (tokenizer.hasMoreElements()) {
                        String value = tokenizer.nextToken();
                        temp.add(value);
                    }
                    map.put("user", temp.get(1));
                    map.put("nice", temp.get(2));
                    map.put("system", temp.get(3));
                    map.put("idle", temp.get(4));
                    map.put("iowait", temp.get(5));
                    map.put("irq", temp.get(6));
                    map.put("softirq", temp.get(7));
                    map.put("stealstolen", temp.get(8));
                    break;
                }
            }
        } catch (Exception e) {
            e.printStackTrace();
        } finally {
            try {
                buffer.close();
                inputs.close();
            } catch (Exception e2) {
                e2.printStackTrace();
            }
        }
        return map;
    }

    /**
     * 功能：Linux 和 Window 内存使用率
     * */
    public static double getMemUsage() {
        if (osName.toLowerCase().contains("windows")
                || osName.toLowerCase().contains("win")) {

            try {
                OperatingSystemMXBean osmxb = (OperatingSystemMXBean) ManagementFactory
                        .getOperatingSystemMXBean();
// 总的物理内存+虚拟内存
                long totalvirtualMemory = osmxb.getTotalSwapSpaceSize();
// 剩余的物理内存
                long freePhysicalMemorySize = osmxb.getFreePhysicalMemorySize();
                Double usage = (Double) (1 - freePhysicalMemorySize * 1.0 / totalvirtualMemory) * 100;
                BigDecimal b1 = new BigDecimal(usage);
                double memoryUsage = b1.setScale(2, BigDecimal.ROUND_HALF_UP).doubleValue();
                return memoryUsage;
            } catch (Exception e) {
                e.printStackTrace();
            }
        } else {
            Map<String, Object> map = new HashMap<String, Object>();
            InputStreamReader inputs = null;
            BufferedReader buffer = null;
            try {
                inputs = new InputStreamReader(new FileInputStream("/proc/meminfo"));
                buffer = new BufferedReader(inputs);
                String line = "";
                while (true) {
                    line = buffer.readLine();
                    if (line == null)
                        break;
                    int beginIndex = 0;
                    int endIndex = line.indexOf(":");
                    if (endIndex != -1) {
                        String key = line.substring(beginIndex, endIndex);
                        beginIndex = endIndex + 1;
                        endIndex = line.length();
                        String memory = line.substring(beginIndex, endIndex);
                        String value = memory.replace("kB", "").trim();
                        map.put(key, value);
                    }
                }

                long memTotal = Long.parseLong(map.get("MemTotal").toString());
                long memFree = Long.parseLong(map.get("MemFree").toString());
                long memused = memTotal - memFree;
                long buffers = Long.parseLong(map.get("Buffers").toString());
                long cached = Long.parseLong(map.get("Cached").toString());

                double usage = (double) (memused - buffers - cached) / memTotal * 100;
                BigDecimal b1 = new BigDecimal(usage);
                double memoryUsage = b1.setScale(2, BigDecimal.ROUND_HALF_UP).doubleValue();
                return memoryUsage;
            } catch (Exception e) {
                e.printStackTrace();
            } finally {
                try {
                    buffer.close();
                    inputs.close();
                } catch (Exception e2) {
                    e2.printStackTrace();
                }
            }

        }
        return 0.0;
    }

    public static float get() {
        float ioUsage = 0.0f;
        Process pro = null;
        Runtime r = Runtime.getRuntime();
        try {
            String command = "iostat -d -x";
            pro = r.exec(command);
            BufferedReader in = new BufferedReader(new InputStreamReader(pro.getInputStream()));
            String line = null;
            int count =  0;
            while((line=in.readLine()) != null){
                if(++count >= 4){
//                  log.info(line);
                    String[] temp = line.split("\\s+");
                    if(temp.length > 1){
                        float util =  Float.parseFloat(temp[temp.length-1]);
                        ioUsage = (ioUsage>util)?ioUsage:util;
                    }
                }
            }
            if(ioUsage > 0){
                ioUsage /= 100;
            }
            in.close();
            pro.destroy();
        } catch (IOException e) {
            StringWriter sw = new StringWriter();
            e.printStackTrace(new PrintWriter(sw));
        }
        return ioUsage;
    }
    /**
     * Window 和Linux 得到磁盘的使用率
     *
     * @return
     * @throws Exception
     */
    public static double getDiskUsage() throws Exception {
        double totalHD = 0;
        double usedHD = 0;
        if (osName.toLowerCase().contains("windows")
                || osName.toLowerCase().contains("win")) {
            long allTotal = 0;
            long allFree = 0;
            for (char c = 'A'; c <= 'Z'; c++) {
                String dirName = c + ":/";
                File win = new File(dirName);
                if (win.exists()) {
                    long total = (long) win.getTotalSpace();
                    long free = (long) win.getFreeSpace();
                    allTotal = allTotal + total;
                    allFree = allFree + free;
                }
            }
            Double precent = (Double) (1 - allFree * 1.0 / allTotal) * 100;
            BigDecimal b1 = new BigDecimal(precent);
            precent = b1.setScale(2, BigDecimal.ROUND_HALF_UP).doubleValue();
            return precent;
        } else {
            Runtime rt = Runtime.getRuntime();
            Process p = rt.exec("df -hl");// df -hl 查看硬盘空间
            BufferedReader in = null;
            try {
                in = new BufferedReader(new InputStreamReader(p.getInputStream()));
                String str = null;
                String[] strArray = null;
                while ((str = in.readLine()) != null) {
                    int m = 0;
                    strArray = str.split(" ");
                    for (String tmp : strArray) {
                        if (tmp.trim().length() == 0)
                            continue;
                        ++m;
                        if (tmp.indexOf("G") != -1) {
                            if (m == 2) {
                                if (!tmp.equals("") && !tmp.equals("0"))
                                    totalHD += Double.parseDouble(tmp.substring(0, tmp.length() - 1)) * 1024;
                            }
                            if (m == 3) {
                                if (!tmp.equals("none") && !tmp.equals("0"))
                                    usedHD += Double.parseDouble(tmp.substring(0, tmp.length() - 1)) * 1024;
                            }
                        }
                        if (tmp.indexOf("M") != -1) {
                            if (m == 2) {
                                if (!tmp.equals("") && !tmp.equals("0"))
                                    totalHD += Double.parseDouble(tmp.substring(0, tmp.length() - 1));
                            }
                            if (m == 3) {
                                if (!tmp.equals("none") && !tmp.equals("0"))
                                    usedHD += Double.parseDouble(tmp.substring(0, tmp.length() - 1));
                            }
                        }
                        if (tmp.indexOf("T") != -1) {
                            if (m == 2) {
                                if (!tmp.equals("") && !tmp.equals("0"))
                                    totalHD += Double.parseDouble(tmp.substring(0, tmp.length() - 1))* 1024*1024;
                            }
                            if (m == 3) {
                                if (!tmp.equals("none") && !tmp.equals("0"))
                                    usedHD += Double.parseDouble(tmp.substring(0, tmp.length() - 1))* 1024*1024;
                            }
                        }
                    }
                }
            } catch (Exception e) {
                e.printStackTrace();
            } finally {
                in.close();
            }
// 保留2位小数
            double precent = (usedHD / totalHD) * 100;
            BigDecimal b1 = new BigDecimal(precent);
            precent = b1.setScale(2, BigDecimal.ROUND_HALF_UP).doubleValue();
            return precent;
        }

    }

    // window读取cpu相关信息
    private static long[] readCpu(final Process proc) {
        long[] retn = new long[2];
        try {
            proc.getOutputStream().close();
            InputStreamReader ir = new InputStreamReader(proc.getInputStream());
            LineNumberReader input = new LineNumberReader(ir);
            String line = input.readLine();
            if (line == null || line.length() < FAULTLENGTH) {
                return null;
            }
            int capidx = line.indexOf("Caption");
            int cmdidx = line.indexOf("CommandLine");
            int rocidx = line.indexOf("ReadOperationCount");
            int umtidx = line.indexOf("UserModeTime");
            int kmtidx = line.indexOf("KernelModeTime");
            int wocidx = line.indexOf("WriteOperationCount");
            long idletime = 0;
            long kneltime = 0;//读取物理设备时间
            long usertime = 0;//执行代码占用时间
            while ((line = input.readLine()) != null) {
                if (line.length() < wocidx) {
                    continue;
                }
// 字段出现顺序：Caption,CommandLine,KernelModeTime,ReadOperationCount
                String caption = substring(line, capidx, cmdidx - 1).trim();
//	System.out.println("caption:"+caption);
                String cmd = substring(line, cmdidx, kmtidx - 1).trim();
//	System.out.println("cmd:"+cmd);
                if (cmd.indexOf("wmic.exe") >= 0) {
                    continue;
                }
                String s1 = substring(line, kmtidx, rocidx - 1).trim();
                String s2 = substring(line, umtidx, wocidx - 1).trim();
                List<String> digitS1 = getDigit(s1);
                List<String> digitS2 = getDigit(s2);

//	System.out.println("s1:"+digitS1.get(0));
//	System.out.println("s2:"+digitS2.get(0));

                if (caption.equals("System Idle Process") || caption.equals("System")) {
                    if (s1.length() > 0) {
                        if (!digitS1.get(0).equals("") && digitS1.get(0) != null) {
                            idletime += Long.valueOf(digitS1.get(0)).longValue();
                        }
                    }
                    if (s2.length() > 0) {
                        if (!digitS2.get(0).equals("") && digitS2.get(0) != null) {
                            idletime += Long.valueOf(digitS2.get(0)).longValue();
                        }
                    }
                    continue;
                }
                if (s1.length() > 0) {
                    if (!digitS1.get(0).equals("") && digitS1.get(0) != null) {
                        kneltime += Long.valueOf(digitS1.get(0)).longValue();
                    }
                }

                if (s2.length() > 0) {
                    if (!digitS2.get(0).equals("") && digitS2.get(0) != null) {
                        kneltime += Long.valueOf(digitS2.get(0)).longValue();
                    }
                }
            }
            retn[0] = idletime;
            retn[1] = kneltime + usertime;

            return retn;
        } catch (Exception ex) {
            ex.printStackTrace();
        } finally {
            try {
                proc.getInputStream().close();
            } catch (Exception e) {
                e.printStackTrace();
            }
        }
        return null;
    }

    /**
     * 从字符串文本中获得数字
     *
     * @param text
     * @return
     */
    private static List<String> getDigit(String text) {
        List<String> digitList = new ArrayList<String>();
        digitList.add(text.replaceAll("\\D", ""));
        return digitList;
    }

    /**
     * 由于String.subString对汉字处理存在问题（把一个汉字视为一个字节)，因此在 包含汉字的字符串时存在隐患，现调整如下：
     *
     * @param src
     * 要截取的字符串
     * @param start_idx
     * 开始坐标（包括该坐标)
     * @param end_idx
     * 截止坐标（包括该坐标）
     * @return
     */
    private static String substring(String src, int start_idx, int end_idx) {
        byte[] b = src.getBytes();
        String tgt = "";
        for (int i = start_idx; i <= end_idx; i++) {
            tgt += (char) b[i];
        }
        return tgt;
    }




}
