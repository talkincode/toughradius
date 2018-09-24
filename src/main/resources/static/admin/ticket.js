if (!window.xspeedercloud.customer.auth_ticket)
    xspeedercloud.customer.auth_ticket={};


xspeedercloud.customer.auth_ticket.loadPage = function(session){
    xspeedercloud.customer.methods.setToolbar("cube","上网日志查询","auth_ticket");
    var tableid = webix.uid();
    var queryid = webix.uid();
    var reloadData = function(){
        $$(tableid).load('/customer/auth/ticket/query');
    };
    webix.ui({
        id:xspeedercloud.customer.panelId,
        css:"main-panel",padding:2,
        rows:[
            {
                padding:5,
                css:"page-toolbar",
                cols:[
                    {view:"icon",icon:"download",width:40},
                    {view:"label",label:"上网日志下载"},
                    {}
                ]
            },
            {
                id: tableid,
                view: "datatable",
                leftSplit: 1,
                rightSplit: 1,
                columns: [
                    { id: "name", header: ["文件名"], width: 240, sort: "string" },
                    { id: "size", header: ["大小"], width: 160, sort: "string", template:function (obj) {
                        return bytesToSize(obj.size);
                    }},
                    { id: "update", header: ["更新时间"],  sort: "string",fillspace:true},
                    { id: "opt", header: '操作', width:120, template: function(obj){
                        return  "<span class='table-btn do_download'><i class='fa fa-download'></i> 日志下载</span> "
                    }},
                    {},
                    {
                        view: "button", type: "icon", width: 70, icon: "refresh", label: "刷新", click: function () {
                            reloadData();
                        }
                    }
                ],
                select: true,
                maxWidth: 2000,
                maxHeight: 2000,
                resizeColumn: true,
                autoWidth: true,
                autoHeight: true,
                url: "/customer/auth/ticket/query",
                pager: "dataPager",
                datafetch: 40,
                loadahead: 15,
                onClick: {
                    do_download:function (e, id) {
                        window.location.href = "/customer/auth/ticket/download?filename="+this.getItem(id).name;
                    }
                }
            },
            {
                paddingY: 3,
                cols: [
                    {
                        id: "dataPager", view: 'pager', master: false, size: 20, group: 5,
                        template: '{common.first()} {common.prev()} {common.pages()} {common.next()} {common.last()} total:#count#'
                    }
                ]
            }
        ]
    },$$(xspeedercloud.customer.pageId),$$(xspeedercloud.customer.panelId));
};
