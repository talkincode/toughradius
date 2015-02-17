#!/usr/bin/env python
#coding=utf-8
from __future__ import division
import math


class Paginator():
    """系统查询分页工具
    """
    def __init__(self, url_func, page=1, total=0, page_size=20):
        self.url_func = url_func
        self.page = 1 if page < 1 else page
        self.total = total
        self.page_size = page_size
        self.page_num = int(math.ceil(self.total / self.page_size)) if self.total > 0 else 0
        self.page_bars = {}
        self.data = ()
        for _page in range(1, self.page_num + 1):
            _index = int(_page / 10)
            if not self.page_bars.has_key(_index):
                self.page_bars[_index] = set([_page])
            else:
                self.page_bars[_index].add(_page)


    def render(self, form_id=None):
        '''
        动态输出html内容
        '''
        page_bar = self.page_bars.get(int(self.page / 10))
        if page_bar is None:
            return ''

        _htmls = []
        if form_id:
            _htmls.append(u'''<script>
                function goto_page(form_id,page){
                    var form=document.getElementById(form_id);
                    var page_input = document.createElement("input");
                    page_input.type="hidden";
                    page_input.name="page";
                    page_input.value=page;
                    form.appendChild(page_input);
                    form.submit();
                }</script>''')
        _htmls.append('<ul class="pagination pull-right">')
        _htmls.append(u'\t<li class="disabled"><a href="#">查询记录数 %s</a></li>' % self.total)
        current_start = self.page
        if current_start == 1:
            _htmls.append(u'\t<li class="disabled"><a href="#">首页</a></li>')
            _htmls.append(u'\t<li class="disabled"><a href="#">&larr; 上一页</a></li>')
        else:
            _htmls.append(u'\t<li><a href="%s">首页</a></li>' % self.url_func(1,form_id))
            _htmls.append(u'\t<li><a href="%s">&larr; 上一页</a></li>' % self.url_func(current_start - 1,form_id))

        for page in page_bar:
            _page_url = self.url_func(page,form_id)
            if page == self.page:
                _htmls.append(u'\t<li class="active"><span>%s <span class="sr-only">{current}</span></span></li>' % page)
            else:
                _htmls.append(u'\t<li><a href="%s">%s</a></li>' % (_page_url, page))



        current_end = self.page
        if current_end == self.page_num:
            _htmls.append(u'\t<li class="disabled"><a href="#">下一页 &rarr;</a></li>')
            _htmls.append(u'\t<li class="disabled"><a href="#">尾页</a></li>')
        else:
            _htmls.append(u'\t<li><a href="%s">下一页 &rarr;</a></li>' % self.url_func(current_end + 1,form_id))
            _htmls.append(u'\t<li><a href="%s">尾页</a></li>' % self.url_func(self.page_num,form_id))

        _htmls.append('</ul>')

        return '\r\n'.join(_htmls)