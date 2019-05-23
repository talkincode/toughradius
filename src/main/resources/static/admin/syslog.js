if (!window.toughradius.admin.syslog)
    toughradius.admin.syslog={};


toughradius.admin.syslog.loadPage = function(session){
    var tableid = webix.uid();
    var queryid = webix.uid();
    var reloadData = function(){
        $$(tableid).define("url", $$(tableid));
        $$(tableid).refresh();
        $$(tableid).clearAll();
        var params = $$(queryid).getValues();
        var args = [];
        for(var k in params){
            args.push(k+"="+params[k]);
        }
        $$(tableid).load('/admin/syslog/query?'+args.join("&"));
    };
    var cview = {
        id:"toughradius.admin.syslog",
        css:"main-panel",padding:10,
        rows:[
            {
                id: queryid,
                css:"query-form",
                view: "form",
                hidden: false,
                paddingX: 10,
                paddingY: 5,
                elementsConfig: {minWidth:180},
                elements: [
                    {
                        type:"space", id:"a1", rows:[{
                            type:"space", padding:0, responsive:"a1", cols:[
                                {view: "datepicker", timepicker:true, name:"startDate",label:"起始时间", stringResult:true, format: "%Y-%m-%d %H:%i"},
                                {view: "datepicker", timepicker:true, name:"endDate",label:"至", stringResult:true,format: "%Y-%m-%d %H:%i"},
                                { view: "text", name: "username", label: "",  placeholder: "用户帐号"},
                                { view: "text", name: "keyword", label: "", placeholder: "内容关键字"},
                                {
                                    cols:[
                                        {view: "richselect",  name: "type", label: "", value: "radiusd", labelWidth:0,width:150,
                                            options: [{ id: "radiusd", value: "RADIUS认证" },
                                                { id: "radiusd_coa", value: "RADIUS_COA" },
                                                { id: "portal", value: "PORTAL认证" },
                                                { id: "api", value: "API接口" },
                                                { id: "error", value: "错误日志" },
                                                { id: "system", value: "系统日志" }]},
                                        {view: "button", label: "查询", type: "icon", icon: "search", borderless: true, width: 55,click:function(){
                                                reloadData();
                                            }},
                                        {view: "button", label: "重置", type: "icon", icon: "refresh", borderless: true, width: 55,click:function(){
                                                $$(queryid).setValues({
                                                    start_date: "",
                                                    end_date: "",
                                                    level: "",
                                                    keyword: ""
                                                });
                                            }},{}
                                    ]
                                }
                            ]
                        }
                    ]}
                ]
            },
            {
                id:tableid,
                view:"treetable",
                subview:{
                    borderless:true,
                    view:"template",
                    height:320,
                    template:"<div style='padding: 5px;'>#msg#</div>"
                },
                on:{
                    onSubViewCreate:function(view, item){
                        item.msg = item.msg.replace("\n","<br>");
                        view.setValues(item);
                    }
                },
                columns:[
                    { id:"name", header:["帐号"], adjust:true, sort: "string",template:"{common.subrow()} #name#"},
                    { id:"type",header:["类型"],sort: "string", adjust:true},
                    { id:"time",header:["时间"], sort: "string",adjust:true},
                    { id:"msg",header:["内容"], fillspace:true},
                    { header: { content: "headerMenu" }, headermenu: false, width: 35 }
                ],
                select: true,
                resizeColumn: true,
                autoWidth: true,
                autoHeight: true,
                url:"/admin/syslog/query",
                pager: "syslog_dataPager",
                datafetch:40,
				loadahead:15,
                ready:function () {
                    reloadData();
                }
            },
            {
                paddingY: 3,
                cols:[
                    {
                        view: "richselect", name: "page_num", label: "每页显示", value: 20,width:130,labelWidth:60,
                        options: [{ id: 20, value: "20" },
                            { id: 50, value: "50" },
                            { id: 100, value: "100" },
                            { id: 500, value: "500" },
                            { id: 1000, value: "1000" }],on: {
                            onChange: function (newv, oldv) {
                                $$("syslog_dataPager").define("size",parseInt(newv));
                                $$(tableid).refresh();
                                reloadData();
                            }
                        }
                    },
                    {
                        id:"syslog_dataPager", view: 'pager', master:false, size: 20, group: 5,
                        template: '{common.first()} {common.prev()} {common.pages()} {common.next()} {common.last()} total:#count#'
                    },{}
                ]
            }
        ]
    };
    toughradius.admin.methods.addTabView("toughradius.admin.syslog","hdd-o","系统日志", cview, true);
    webix.extend($$(tableid), webix.ProgressBar);
};




