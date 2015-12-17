#!/usr/bin/env python
# coding:utf-8

from hashlib import md5

import cyclone.auth
import cyclone.escape
import cyclone.web

from toughradius.common import utils
from toughradius.console.admin.base import BaseHandler
from toughradius.common.permit import permit
from toughradius.console import models
from toughradius.console.admin.system import operator_form
from toughradius.console.admin.system.operator_form import opr_status_dict
from toughradius.common.settings import * 


@permit.route(r"/operator", u"操作员管理", MenuSys, order=3.0000, is_menu=True)
class OperatorHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        self.render("operator_list.html",
                      operator_list=self.db.query(models.TrOperator),opr_status=opr_status_dict)


@permit.route(r"/operator/add", u"操作员新增", MenuSys, order=3.0001)
class AddHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        nodes = [(n.node_name, n.node_desc) for n in self.db.query(models.TrNode)]
        products = [(p.id,p.product_name) for p in self.db.query(models.TrProduct)  ]
        self.render("opr_form.html", form=operator_form.operator_add_form(nodes,products),rules=[])

    @cyclone.web.authenticated
    def post(self):
        nodes = [(n.node_name, n.node_desc) for n in self.db.query(models.TrNode)]
        products = [(p.id,p.product_name) for p in self.db.query(models.TrProduct)]
        form = operator_form.operator_add_form(nodes,products)
        if not form.validates(source=self.get_params()):
            return self.render("base_form.html", form=form)
        if self.db.query(models.TrOperator.id).filter_by(operator_name=form.d.operator_name).count() > 0:
            return self.render("base_form.html", form=form, msg=u"操作员已经存在")
        operator = models.TrOperator()
        operator.operator_name = form.d.operator_name
        operator.operator_pass = md5(form.d.operator_pass.encode()).hexdigest()
        operator.operator_type = 1
        operator.operator_desc = form.d.operator_desc
        operator.operator_status = form.d.operator_status
        self.db.add(operator)

        self.add_oplog(u'新增操作员信息:%s' % utils.safeunicode(operator.operator_name))

        for node in self.get_arguments("operator_nodes"):
            onode = models.TrOperatorNodes()
            onode.operator_name = form.d.operator_name
            onode.node_name = node
            self.db.add(onode)

        for product_id in self.get_arguments("operator_products"):
            oproduct = models.TrOperatorProducts()
            oproduct.operator_name = form.d.operator_name
            oproduct.product_id = product_id
            self.db.add(oproduct)

        for path in self.get_arguments("rule_item"):
            item = permit.get_route(path)
            if not item: continue
            rule = models.TrOperatorRule()
            rule.operator_name = operator.operator_name
            rule.rule_name = item['name']
            rule.rule_path = item['path']
            rule.rule_category = item['category']
            self.db.add(rule)

        self.db.commit()

        for rule in self.db.query(models.TrOperatorRule).filter_by(operator_name=operator.operator_name):
            permit.bind_opr(rule.operator_name, rule.rule_path)

        self.redirect("/operator",permanent=False)

@permit.route(r"/operator/update", u"操作员修改", MenuSys, order=3.0002)
class UpdateHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        operator_id = self.get_argument("operator_id")
        opr = self.db.query(models.TrOperator).get(operator_id)
        nodes = [(n.node_name, n.node_desc) for n in self.db.query(models.TrNode)]
        products = [(p.id,p.product_name) for p in self.db.query(models.TrProduct)]
        form = operator_form.operator_update_form(nodes,products)
        form.fill(self.db.query(models.TrOperator).get(operator_id))
        form.operator_pass.set_value('')

        onodes = self.db.query(models.TrOperatorNodes).filter_by(operator_name=form.d.operator_name)
        oproducts = self.db.query(models.TrOperatorProducts).filter_by(operator_name=form.d.operator_name)
        form.operator_products.set_value([int(p.product_id) for p in oproducts])
        form.operator_nodes.set_value([ond.node_name for ond in onodes])
        

        rules = self.db.query(models.TrOperatorRule.rule_path).filter_by(operator_name=opr.operator_name)
        rules = [r[0] for r in rules]
        return self.render("opr_form.html", form=form, rules=rules)

    @cyclone.web.authenticated
    def post(self):
        nodes = [(n.node_name, n.node_desc) for n in self.db.query(models.TrNode)]
        products = [(p.id,p.product_name) for p in self.db.query(models.TrProduct)]
        form = operator_form.operator_update_form(nodes,products)
        if not form.validates(source=self.get_params()):
            rules = self.db.query(models.TrOperatorRule.rule_path).filter_by(operator_name=form.d.operator_name)
            rules = [r[0] for r in rules]
            return self.render("base_form.html", form=form,rules=rules)
        operator = self.db.query(models.TrOperator).get(form.d.id)
        if form.d.operator_pass:
            operator.operator_pass = md5(form.d.operator_pass.encode()).hexdigest()
        operator.operator_desc = form.d.operator_desc
        operator.operator_status = form.d.operator_status

        self.db.query(models.TrOperatorNodes).filter_by(operator_name=operator.operator_name).delete()
        for node in self.get_arguments("operator_nodes"):
            onode = models.TrOperatorNodes()
            onode.operator_name = form.d.operator_name
            onode.node_name = node
            self.db.add(onode)

        self.db.query(models.TrOperatorProducts).filter_by(operator_name=operator.operator_name).delete()
        for product_id in self.get_arguments("operator_products"):
            oproduct = models.TrOperatorProducts()
            oproduct.operator_name = form.d.operator_name
            oproduct.product_id = product_id
            self.db.add(oproduct)

        self.add_oplog(u'修改操作员%s信息' % utils.safeunicode(operator.operator_name))

        # update rules
        self.db.query(models.TrOperatorRule).filter_by(operator_name=operator.operator_name).delete()

        for path in self.get_arguments("rule_item"):
            item = permit.get_route(path)
            if not item: continue
            rule = models.TrOperatorRule()
            rule.operator_name = operator.operator_name
            rule.rule_name = item['name']
            rule.rule_path = item['path']
            rule.rule_category = item['category']
            self.db.add(rule)

        permit.unbind_opr(operator.operator_name)

        self.db.commit()

        for rule in self.db.query(models.TrOperatorRule).filter_by(operator_name=operator.operator_name):
            permit.bind_opr(rule.operator_name, rule.rule_path)

        self.redirect("/operator",permanent=False)

@permit.route(r"/operator/delete", u"操作员删除", MenuSys, order=3.0003)
class DeleteHandler(BaseHandler):

    @cyclone.web.authenticated
    def get(self):
        operator_id = self.get_argument("operator_id")
        opr = self.db.query(models.TrOperator).get(operator_id)
        self.db.query(models.TrOperatorRule).filter_by(operator_name=opr.operator_name).delete()
        self.db.query(models.TrOperator).filter_by(id=operator_id).delete()

        self.add_oplog(u'删除操作员%s信息' % utils.safeunicode(opr.operator_name))
        self.db.commit()
        self.redirect("/operator",permanent=False)




