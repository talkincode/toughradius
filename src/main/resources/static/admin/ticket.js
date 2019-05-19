if (!window.toughradius.admin.ticket)
    toughradius.admin.ticket={};


toughradius.admin.ticket.loadPage = function(session){
    var tableid = webix.uid();
    var queryid = webix.uid();
    var reloadData = function(node_id){
        $$(tableid).define("url", $$(tableid));
        $$(tableid).refresh();
        $$(tableid).clearAll();
        var params = $$(queryid).getValues();
        var args = [];
        for(var k in params){
            args.push(k+"="+params[k]);
        }
        $$(tableid).load('/admin/ticket/query?'+args.join("&"));
    };
    var cview = {
        id:"toughradius.admin.ticket",
        css:"main-panel",padding:10,
        rows: [
            {
                view: "toolbar",
                css: "page-toolbar",
                paddingX:10,
                cols: [
                    {view:"label",css: "form-desc",label:"注意：查询时间跨度在30天以内，最大返回10000条记录，缩小查询范围可提升查询速度"},
                    { },
                    {
                        view: "button", type: "icon", width: 70, icon: "refresh", label: "刷新", click: function () {
                            reloadData();
                        }
                    }
                ]
            },
            {
                id: queryid,
                css:"query-form",
                view: "form",
                hidden: false,
                paddingX: 10,
                paddingY: 5,
                elementsConfig: {minWidth:150},
                elements: [
                    {
                       type:"space", id:"a1", rows:[{
                         type:"space", padding:0, responsive:"a1", cols:[
                            { view: "datepicker", name: "startDate", label: "上线时间",stringResult:true, timepicker: true, format: "%Y-%m-%d %h:%i"},
                            { view: "datepicker", name: "endDate", label: "至", value:new Date(),stringResult:true, timepicker: true, format: "%Y-%m-%d %h:%i" },
                            {view: "text", name: "keyword", label: "关键字",  placeholder: "帐号 / IP地址 / MAC地址 .."},
                            {
                                cols:[

                                    {view: "button", label: "查询", type: "icon", icon: "search", borderless: true, width:60, click:function(){
                                            reloadData();
                                        }},
                                    {view: "button", label: "重置", type: "icon", icon: "refresh", borderless: true,width:60, click:function(){
                                            $$(queryid).setValues({
                                                startDate: "",
                                                endDate: "",
                                                keyword: "",
                                                area_id: "",
                                                zone_id: "",
                                                status: ""
                                            });
                                        }},{}
                                ]
                            }
                         ]}
                       ]
                    }
                ]
            },
            {
                id: tableid,
                view: "datatable",
                rightSplit: 1,
                columns: [
                    { id: "username", header: ["用户名"], sort: "string" , adjust:true},
                    { id: "acctSessionId", header: ["会话ID"],  sort: "string", hidden: true , adjust:true },
                    { id: "nasId", header: ["BRAS 标识"], sort: "string"  , adjust:true},
                    { id: "acctStartTime", header: ["上线时间"],  sort: "string" , adjust:true },
                    { id: "acctStopTime", header: ["下线时间"], sort: "string" , adjust:true },
                    { id: "nasAddr", header: ["BRAS IP"], sort: "string" , adjust:true },
                    { id: "framedIpaddr", header: ["用户 IP"],  sort: "string"  , adjust:true},
                    { id: "macAddr", header: ["用户 Mac"], sort: "string"  , adjust:true},
                    { id: "nasPortId", header: ["端口信息"],  sort: "string" , adjust:true},
                    {
                        id: "acctInputTotal", header: ["上传"],  sort: "int" , adjust:true, template: function (obj) {
                            return bytesToSize(obj.acctInputTotal);
                        }
                    },
                    {
                        id: "acctOutputTotal", header: ["下载"], sort: "int" , adjust:true, template: function (obj) {
                            return bytesToSize(obj.acctOutputTotal);
                        }
                    },
                    { id: "acctInputPackets", header: ["上行数据包"],  sort: "string", adjust:true},
                    { id: "acctOutputPackets", header: ["下行数据包"], sort: "string", adjust:true},
                    { id: "_", header: [""],   fillspace:true},
                    { header: { content: "headerMenu" }, headermenu: false, width: 35 }
                ],
                select: true,
                resizeColumn: true,
                autoWidth: true,
                autoHeight: true,
                url: "/admin/ticket/query",
                pager: "ticket_dataPager",
                datafetch: 40,
                loadahead: 15,
                on: {}
            },
            {
                paddingY: 3,
                cols: [
                    {
                        view: "richselect", name: "page_num", label: "每页显示", value: 20,width:130,labelWidth:60,
                        options: [{ id: 20, value: "20" },
                            { id: 50, value: "50" },
                            { id: 100, value: "100" },
                            { id: 500, value: "500" },
                            { id: 1000, value: "1000" }],on: {
                            onChange: function (newv, oldv) {
                                $$("ticket_dataPager").define("size",parseInt(newv));
                                $$(tableid).refresh();
                                reloadData();
                            }
                        }
                    },
                    {
                        id: "ticket_dataPager", view: 'pager', master: false, size: 20, group: 5,
                        template: '{common.first()} {common.prev()} {common.pages()} {common.next()} {common.last()} total:#count#'
                    },{}
                ]
            }
        ]
    };
    toughradius.admin.methods.addTabView("toughradius.admin.ticket","hdd-o","上网日志", cview, true);
    webix.extend($$(tableid), webix.ProgressBar);
};
