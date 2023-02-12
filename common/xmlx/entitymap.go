// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package xmlx

/*
	These routines offer conversions between xml entities and their respective
	unicode representations.

	eg:
		&#9827;  -> ♣   -> &#9827;
		&pi;     -> π   -> &#960;

	Note that named entities are case sensitive.
	"&acirc;" (â) is not the same as "&Acirc;" (Â).
*/

import (
	"fmt"
	"regexp"
	"strconv"
	"unicode/utf8"
)

var reg_entnumeric = regexp.MustCompile("^&#[0-9]+;$")
var reg_entnamed = regexp.MustCompile("^&[a-zA-Z]+;$")

// Converts a single numerical html entity to a regular Go utf8-token.
func EntityToUtf8(entity string) string {
	var ok bool
	if ok = reg_entnamed.MatchString(entity); ok {
		return namedEntityToUtf8(entity[1 : len(entity)-1])
	}

	if ok = reg_entnumeric.MatchString(entity); !ok {
		return "&amp;" + entity[2:len(entity)-1] + ";"
	}

	var err error
	var num int

	entity = entity[2 : len(entity)-1]
	if num, err = strconv.Atoi(entity); err != nil {
		return "&amp;#" + entity + ";"
	}

	var arr [4]byte
	if size := utf8.EncodeRune(arr[:], rune(num)); size == 0 {
		return "&amp;#" + entity + ";"
	}

	return string(arr[:])
}

// Converts a single Go utf8-token to a Html entity.
func Utf8ToEntity(entity string) string {
	if rune, size := utf8.DecodeRuneInString(entity); size != 0 {
		return fmt.Sprintf("&#%d;", rune)
	}
	return entity
}

/*
	http://www.w3.org/TR/html4/sgml/entities.html

	Portions © International Organization for Standardization 1986
	Permission to copy in any form is granted for use with
	conforming SGML systems and applications as defined in
	ISO 8879, provided this notice is included in all copies.

	Fills the supplied map with html entities mapped to their Go utf8
	equivalents. This map can be assigned to xml.Parser.Entity
	It will be used to map non-standard xml entities to a proper value.
	If the parser encounters any unknown entities, it will throw a syntax
	error and abort the parsing. Hence the ability to supply this map.
*/
func loadNonStandardEntities(em map[string]string) {
	em["pi"] = "\u03c0"
	em["nabla"] = "\u2207"
	em["isin"] = "\u2208"
	em["loz"] = "\u25ca"
	em["prop"] = "\u221d"
	em["para"] = "\u00b6"
	em["Aring"] = "\u00c5"
	em["euro"] = "\u20ac"
	em["sup3"] = "\u00b3"
	em["sup2"] = "\u00b2"
	em["sup1"] = "\u00b9"
	em["prod"] = "\u220f"
	em["gamma"] = "\u03b3"
	em["perp"] = "\u22a5"
	em["lfloor"] = "\u230a"
	em["fnof"] = "\u0192"
	em["frasl"] = "\u2044"
	em["rlm"] = "\u200f"
	em["omega"] = "\u03c9"
	em["part"] = "\u2202"
	em["euml"] = "\u00eb"
	em["Kappa"] = "\u039a"
	em["nbsp"] = "\u00a0"
	em["Eacute"] = "\u00c9"
	em["brvbar"] = "\u00a6"
	em["otimes"] = "\u2297"
	em["ndash"] = "\u2013"
	em["thinsp"] = "\u2009"
	em["nu"] = "\u03bd"
	em["Upsilon"] = "\u03a5"
	em["upsih"] = "\u03d2"
	em["raquo"] = "\u00bb"
	em["yacute"] = "\u00fd"
	em["delta"] = "\u03b4"
	em["eth"] = "\u00f0"
	em["supe"] = "\u2287"
	em["ne"] = "\u2260"
	em["ni"] = "\u220b"
	em["eta"] = "\u03b7"
	em["uArr"] = "\u21d1"
	em["image"] = "\u2111"
	em["asymp"] = "\u2248"
	em["oacute"] = "\u00f3"
	em["rarr"] = "\u2192"
	em["emsp"] = "\u2003"
	em["acirc"] = "\u00e2"
	em["shy"] = "\u00ad"
	em["yuml"] = "\u00ff"
	em["acute"] = "\u00b4"
	em["int"] = "\u222b"
	em["ccedil"] = "\u00e7"
	em["Acirc"] = "\u00c2"
	em["Ograve"] = "\u00d2"
	em["times"] = "\u00d7"
	em["weierp"] = "\u2118"
	em["Tau"] = "\u03a4"
	em["omicron"] = "\u03bf"
	em["lt"] = "\u003c"
	em["Mu"] = "\u039c"
	em["Ucirc"] = "\u00db"
	em["sub"] = "\u2282"
	em["le"] = "\u2264"
	em["sum"] = "\u2211"
	em["sup"] = "\u2283"
	em["lrm"] = "\u200e"
	em["frac34"] = "\u00be"
	em["Iota"] = "\u0399"
	em["Ugrave"] = "\u00d9"
	em["THORN"] = "\u00de"
	em["rsaquo"] = "\u203a"
	em["not"] = "\u00ac"
	em["sigma"] = "\u03c3"
	em["iuml"] = "\u00ef"
	em["epsilon"] = "\u03b5"
	em["spades"] = "\u2660"
	em["theta"] = "\u03b8"
	em["divide"] = "\u00f7"
	em["Atilde"] = "\u00c3"
	em["uacute"] = "\u00fa"
	em["Rho"] = "\u03a1"
	em["trade"] = "\u2122"
	em["chi"] = "\u03c7"
	em["agrave"] = "\u00e0"
	em["or"] = "\u2228"
	em["circ"] = "\u02c6"
	em["middot"] = "\u00b7"
	em["plusmn"] = "\u00b1"
	em["aring"] = "\u00e5"
	em["lsquo"] = "\u2018"
	em["Yacute"] = "\u00dd"
	em["oline"] = "\u203e"
	em["copy"] = "\u00a9"
	em["icirc"] = "\u00ee"
	em["lowast"] = "\u2217"
	em["Oacute"] = "\u00d3"
	em["aacute"] = "\u00e1"
	em["oplus"] = "\u2295"
	em["crarr"] = "\u21b5"
	em["thetasym"] = "\u03d1"
	em["Beta"] = "\u0392"
	em["laquo"] = "\u00ab"
	em["rang"] = "\u232a"
	em["tilde"] = "\u02dc"
	em["Uuml"] = "\u00dc"
	em["zwj"] = "\u200d"
	em["mu"] = "\u03bc"
	em["Ccedil"] = "\u00c7"
	em["infin"] = "\u221e"
	em["ouml"] = "\u00f6"
	em["rfloor"] = "\u230b"
	em["pound"] = "\u00a3"
	em["szlig"] = "\u00df"
	em["thorn"] = "\u00fe"
	em["forall"] = "\u2200"
	em["piv"] = "\u03d6"
	em["rdquo"] = "\u201d"
	em["frac12"] = "\u00bd"
	em["frac14"] = "\u00bc"
	em["Ocirc"] = "\u00d4"
	em["Ecirc"] = "\u00ca"
	em["kappa"] = "\u03ba"
	em["Euml"] = "\u00cb"
	em["minus"] = "\u2212"
	em["cong"] = "\u2245"
	em["hellip"] = "\u2026"
	em["equiv"] = "\u2261"
	em["cent"] = "\u00a2"
	em["Uacute"] = "\u00da"
	em["darr"] = "\u2193"
	em["Eta"] = "\u0397"
	em["sbquo"] = "\u201a"
	em["rArr"] = "\u21d2"
	em["igrave"] = "\u00ec"
	em["uml"] = "\u00a8"
	em["lambda"] = "\u03bb"
	em["oelig"] = "\u0153"
	em["harr"] = "\u2194"
	em["ang"] = "\u2220"
	em["clubs"] = "\u2663"
	em["and"] = "\u2227"
	em["permil"] = "\u2030"
	em["larr"] = "\u2190"
	em["Yuml"] = "\u0178"
	em["cup"] = "\u222a"
	em["Xi"] = "\u039e"
	em["Alpha"] = "\u0391"
	em["phi"] = "\u03c6"
	em["ucirc"] = "\u00fb"
	em["oslash"] = "\u00f8"
	em["rsquo"] = "\u2019"
	em["AElig"] = "\u00c6"
	em["mdash"] = "\u2014"
	em["psi"] = "\u03c8"
	em["eacute"] = "\u00e9"
	em["otilde"] = "\u00f5"
	em["yen"] = "\u00a5"
	em["gt"] = "\u003e"
	em["Iuml"] = "\u00cf"
	em["Prime"] = "\u2033"
	em["Chi"] = "\u03a7"
	em["ge"] = "\u2265"
	em["reg"] = "\u00ae"
	em["hearts"] = "\u2665"
	em["auml"] = "\u00e4"
	em["Agrave"] = "\u00c0"
	em["sect"] = "\u00a7"
	em["sube"] = "\u2286"
	em["sigmaf"] = "\u03c2"
	em["Gamma"] = "\u0393"
	em["amp"] = "\u0026"
	em["ensp"] = "\u2002"
	em["ETH"] = "\u00d0"
	em["Igrave"] = "\u00cc"
	em["Omega"] = "\u03a9"
	em["Lambda"] = "\u039b"
	em["Omicron"] = "\u039f"
	em["there4"] = "\u2234"
	em["ntilde"] = "\u00f1"
	em["xi"] = "\u03be"
	em["dagger"] = "\u2020"
	em["egrave"] = "\u00e8"
	em["Delta"] = "\u0394"
	em["OElig"] = "\u0152"
	em["diams"] = "\u2666"
	em["ldquo"] = "\u201c"
	em["radic"] = "\u221a"
	em["Oslash"] = "\u00d8"
	em["Ouml"] = "\u00d6"
	em["lceil"] = "\u2308"
	em["uarr"] = "\u2191"
	em["atilde"] = "\u00e3"
	em["iquest"] = "\u00bf"
	em["lsaquo"] = "\u2039"
	em["Epsilon"] = "\u0395"
	em["iacute"] = "\u00ed"
	em["cap"] = "\u2229"
	em["deg"] = "\u00b0"
	em["Otilde"] = "\u00d5"
	em["zeta"] = "\u03b6"
	em["ocirc"] = "\u00f4"
	em["scaron"] = "\u0161"
	em["ecirc"] = "\u00ea"
	em["ordm"] = "\u00ba"
	em["tau"] = "\u03c4"
	em["Auml"] = "\u00c4"
	em["dArr"] = "\u21d3"
	em["ordf"] = "\u00aa"
	em["alefsym"] = "\u2135"
	em["notin"] = "\u2209"
	em["Pi"] = "\u03a0"
	em["sdot"] = "\u22c5"
	em["upsilon"] = "\u03c5"
	em["iota"] = "\u03b9"
	em["hArr"] = "\u21d4"
	em["Sigma"] = "\u03a3"
	em["lang"] = "\u2329"
	em["curren"] = "\u00a4"
	em["Theta"] = "\u0398"
	em["lArr"] = "\u21d0"
	em["Phi"] = "\u03a6"
	em["Nu"] = "\u039d"
	em["rho"] = "\u03c1"
	em["alpha"] = "\u03b1"
	em["iexcl"] = "\u00a1"
	em["micro"] = "\u00b5"
	em["cedil"] = "\u00b8"
	em["Ntilde"] = "\u00d1"
	em["Psi"] = "\u03a8"
	em["Dagger"] = "\u2021"
	em["Egrave"] = "\u00c8"
	em["Icirc"] = "\u00ce"
	em["nsub"] = "\u2284"
	em["bdquo"] = "\u201e"
	em["empty"] = "\u2205"
	em["aelig"] = "\u00e6"
	em["ograve"] = "\u00f2"
	em["macr"] = "\u00af"
	em["Zeta"] = "\u0396"
	em["beta"] = "\u03b2"
	em["sim"] = "\u223c"
	em["uuml"] = "\u00fc"
	em["Aacute"] = "\u00c1"
	em["Iacute"] = "\u00cd"
	em["exist"] = "\u2203"
	em["prime"] = "\u2032"
	em["rceil"] = "\u2309"
	em["real"] = "\u211c"
	em["zwnj"] = "\u200c"
	em["bull"] = "\u2022"
	em["quot"] = "\u0022"
	em["Scaron"] = "\u0160"
	em["ugrave"] = "\u00f9"
}

/*
	http://www.w3.org/TR/html4/sgml/entities.html

	Portions © International Organization for Standardization 1986
	Permission to copy in any form is granted for use with
	conforming SGML systems and applications as defined in
	ISO 8879, provided this notice is included in all copies.
*/
func namedEntityToUtf8(name string) string {
	switch name {
	case "pi":
		return "\u03c0"
	case "nabla":
		return "\u2207"
	case "isin":
		return "\u2208"
	case "loz":
		return "\u25ca"
	case "prop":
		return "\u221d"
	case "para":
		return "\u00b6"
	case "Aring":
		return "\u00c5"
	case "euro":
		return "\u20ac"
	case "sup3":
		return "\u00b3"
	case "sup2":
		return "\u00b2"
	case "sup1":
		return "\u00b9"
	case "prod":
		return "\u220f"
	case "gamma":
		return "\u03b3"
	case "perp":
		return "\u22a5"
	case "lfloor":
		return "\u230a"
	case "fnof":
		return "\u0192"
	case "frasl":
		return "\u2044"
	case "rlm":
		return "\u200f"
	case "omega":
		return "\u03c9"
	case "part":
		return "\u2202"
	case "euml":
		return "\u00eb"
	case "Kappa":
		return "\u039a"
	case "nbsp":
		return "\u00a0"
	case "Eacute":
		return "\u00c9"
	case "brvbar":
		return "\u00a6"
	case "otimes":
		return "\u2297"
	case "ndash":
		return "\u2013"
	case "thinsp":
		return "\u2009"
	case "nu":
		return "\u03bd"
	case "Upsilon":
		return "\u03a5"
	case "upsih":
		return "\u03d2"
	case "raquo":
		return "\u00bb"
	case "yacute":
		return "\u00fd"
	case "delta":
		return "\u03b4"
	case "eth":
		return "\u00f0"
	case "supe":
		return "\u2287"
	case "ne":
		return "\u2260"
	case "ni":
		return "\u220b"
	case "eta":
		return "\u03b7"
	case "uArr":
		return "\u21d1"
	case "image":
		return "\u2111"
	case "asymp":
		return "\u2248"
	case "oacute":
		return "\u00f3"
	case "rarr":
		return "\u2192"
	case "emsp":
		return "\u2003"
	case "acirc":
		return "\u00e2"
	case "shy":
		return "\u00ad"
	case "yuml":
		return "\u00ff"
	case "acute":
		return "\u00b4"
	case "int":
		return "\u222b"
	case "ccedil":
		return "\u00e7"
	case "Acirc":
		return "\u00c2"
	case "Ograve":
		return "\u00d2"
	case "times":
		return "\u00d7"
	case "weierp":
		return "\u2118"
	case "Tau":
		return "\u03a4"
	case "omicron":
		return "\u03bf"
	case "lt":
		return "\u003c"
	case "Mu":
		return "\u039c"
	case "Ucirc":
		return "\u00db"
	case "sub":
		return "\u2282"
	case "le":
		return "\u2264"
	case "sum":
		return "\u2211"
	case "sup":
		return "\u2283"
	case "lrm":
		return "\u200e"
	case "frac34":
		return "\u00be"
	case "Iota":
		return "\u0399"
	case "Ugrave":
		return "\u00d9"
	case "THORN":
		return "\u00de"
	case "rsaquo":
		return "\u203a"
	case "not":
		return "\u00ac"
	case "sigma":
		return "\u03c3"
	case "iuml":
		return "\u00ef"
	case "epsilon":
		return "\u03b5"
	case "spades":
		return "\u2660"
	case "theta":
		return "\u03b8"
	case "divide":
		return "\u00f7"
	case "Atilde":
		return "\u00c3"
	case "uacute":
		return "\u00fa"
	case "Rho":
		return "\u03a1"
	case "trade":
		return "\u2122"
	case "chi":
		return "\u03c7"
	case "agrave":
		return "\u00e0"
	case "or":
		return "\u2228"
	case "circ":
		return "\u02c6"
	case "middot":
		return "\u00b7"
	case "plusmn":
		return "\u00b1"
	case "aring":
		return "\u00e5"
	case "lsquo":
		return "\u2018"
	case "Yacute":
		return "\u00dd"
	case "oline":
		return "\u203e"
	case "copy":
		return "\u00a9"
	case "icirc":
		return "\u00ee"
	case "lowast":
		return "\u2217"
	case "Oacute":
		return "\u00d3"
	case "aacute":
		return "\u00e1"
	case "oplus":
		return "\u2295"
	case "crarr":
		return "\u21b5"
	case "thetasym":
		return "\u03d1"
	case "Beta":
		return "\u0392"
	case "laquo":
		return "\u00ab"
	case "rang":
		return "\u232a"
	case "tilde":
		return "\u02dc"
	case "Uuml":
		return "\u00dc"
	case "zwj":
		return "\u200d"
	case "mu":
		return "\u03bc"
	case "Ccedil":
		return "\u00c7"
	case "infin":
		return "\u221e"
	case "ouml":
		return "\u00f6"
	case "rfloor":
		return "\u230b"
	case "pound":
		return "\u00a3"
	case "szlig":
		return "\u00df"
	case "thorn":
		return "\u00fe"
	case "forall":
		return "\u2200"
	case "piv":
		return "\u03d6"
	case "rdquo":
		return "\u201d"
	case "frac12":
		return "\u00bd"
	case "frac14":
		return "\u00bc"
	case "Ocirc":
		return "\u00d4"
	case "Ecirc":
		return "\u00ca"
	case "kappa":
		return "\u03ba"
	case "Euml":
		return "\u00cb"
	case "minus":
		return "\u2212"
	case "cong":
		return "\u2245"
	case "hellip":
		return "\u2026"
	case "equiv":
		return "\u2261"
	case "cent":
		return "\u00a2"
	case "Uacute":
		return "\u00da"
	case "darr":
		return "\u2193"
	case "Eta":
		return "\u0397"
	case "sbquo":
		return "\u201a"
	case "rArr":
		return "\u21d2"
	case "igrave":
		return "\u00ec"
	case "uml":
		return "\u00a8"
	case "lambda":
		return "\u03bb"
	case "oelig":
		return "\u0153"
	case "harr":
		return "\u2194"
	case "ang":
		return "\u2220"
	case "clubs":
		return "\u2663"
	case "and":
		return "\u2227"
	case "permil":
		return "\u2030"
	case "larr":
		return "\u2190"
	case "Yuml":
		return "\u0178"
	case "cup":
		return "\u222a"
	case "Xi":
		return "\u039e"
	case "Alpha":
		return "\u0391"
	case "phi":
		return "\u03c6"
	case "ucirc":
		return "\u00fb"
	case "oslash":
		return "\u00f8"
	case "rsquo":
		return "\u2019"
	case "AElig":
		return "\u00c6"
	case "mdash":
		return "\u2014"
	case "psi":
		return "\u03c8"
	case "eacute":
		return "\u00e9"
	case "otilde":
		return "\u00f5"
	case "yen":
		return "\u00a5"
	case "gt":
		return "\u003e"
	case "Iuml":
		return "\u00cf"
	case "Prime":
		return "\u2033"
	case "Chi":
		return "\u03a7"
	case "ge":
		return "\u2265"
	case "reg":
		return "\u00ae"
	case "hearts":
		return "\u2665"
	case "auml":
		return "\u00e4"
	case "Agrave":
		return "\u00c0"
	case "sect":
		return "\u00a7"
	case "sube":
		return "\u2286"
	case "sigmaf":
		return "\u03c2"
	case "Gamma":
		return "\u0393"
	case "amp":
		return "\u0026"
	case "ensp":
		return "\u2002"
	case "ETH":
		return "\u00d0"
	case "Igrave":
		return "\u00cc"
	case "Omega":
		return "\u03a9"
	case "Lambda":
		return "\u039b"
	case "Omicron":
		return "\u039f"
	case "there4":
		return "\u2234"
	case "ntilde":
		return "\u00f1"
	case "xi":
		return "\u03be"
	case "dagger":
		return "\u2020"
	case "egrave":
		return "\u00e8"
	case "Delta":
		return "\u0394"
	case "OElig":
		return "\u0152"
	case "diams":
		return "\u2666"
	case "ldquo":
		return "\u201c"
	case "radic":
		return "\u221a"
	case "Oslash":
		return "\u00d8"
	case "Ouml":
		return "\u00d6"
	case "lceil":
		return "\u2308"
	case "uarr":
		return "\u2191"
	case "atilde":
		return "\u00e3"
	case "iquest":
		return "\u00bf"
	case "lsaquo":
		return "\u2039"
	case "Epsilon":
		return "\u0395"
	case "iacute":
		return "\u00ed"
	case "cap":
		return "\u2229"
	case "deg":
		return "\u00b0"
	case "Otilde":
		return "\u00d5"
	case "zeta":
		return "\u03b6"
	case "ocirc":
		return "\u00f4"
	case "scaron":
		return "\u0161"
	case "ecirc":
		return "\u00ea"
	case "ordm":
		return "\u00ba"
	case "tau":
		return "\u03c4"
	case "Auml":
		return "\u00c4"
	case "dArr":
		return "\u21d3"
	case "ordf":
		return "\u00aa"
	case "alefsym":
		return "\u2135"
	case "notin":
		return "\u2209"
	case "Pi":
		return "\u03a0"
	case "sdot":
		return "\u22c5"
	case "upsilon":
		return "\u03c5"
	case "iota":
		return "\u03b9"
	case "hArr":
		return "\u21d4"
	case "Sigma":
		return "\u03a3"
	case "lang":
		return "\u2329"
	case "curren":
		return "\u00a4"
	case "Theta":
		return "\u0398"
	case "lArr":
		return "\u21d0"
	case "Phi":
		return "\u03a6"
	case "Nu":
		return "\u039d"
	case "rho":
		return "\u03c1"
	case "alpha":
		return "\u03b1"
	case "iexcl":
		return "\u00a1"
	case "micro":
		return "\u00b5"
	case "cedil":
		return "\u00b8"
	case "Ntilde":
		return "\u00d1"
	case "Psi":
		return "\u03a8"
	case "Dagger":
		return "\u2021"
	case "Egrave":
		return "\u00c8"
	case "Icirc":
		return "\u00ce"
	case "nsub":
		return "\u2284"
	case "bdquo":
		return "\u201e"
	case "empty":
		return "\u2205"
	case "aelig":
		return "\u00e6"
	case "ograve":
		return "\u00f2"
	case "macr":
		return "\u00af"
	case "Zeta":
		return "\u0396"
	case "beta":
		return "\u03b2"
	case "sim":
		return "\u223c"
	case "uuml":
		return "\u00fc"
	case "Aacute":
		return "\u00c1"
	case "Iacute":
		return "\u00cd"
	case "exist":
		return "\u2203"
	case "prime":
		return "\u2032"
	case "rceil":
		return "\u2309"
	case "real":
		return "\u211c"
	case "zwnj":
		return "\u200c"
	case "bull":
		return "\u2022"
	case "quot":
		return "\u0022"
	case "Scaron":
		return "\u0160"
	case "ugrave":
		return "\u00f9"
	}
	return "&amp;" + name + ";"
}
