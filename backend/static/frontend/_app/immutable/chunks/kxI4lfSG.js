import{e as x}from"./C7gi3OLk.js";const M="[A-Za-z$_][0-9A-Za-z$_]*",K=["as","in","of","if","for","while","finally","var","new","function","do","return","void","else","break","catch","instanceof","with","throw","case","default","try","switch","continue","typeof","delete","let","yield","const","class","debugger","async","await","static","import","from","export","extends","using"],W=["true","false","null","undefined","NaN","Infinity"],L=["Object","Function","Boolean","Symbol","Math","Date","Number","BigInt","String","RegExp","Array","Float32Array","Float64Array","Int8Array","Uint8Array","Uint8ClampedArray","Int16Array","Int32Array","Uint16Array","Uint32Array","BigInt64Array","BigUint64Array","Set","Map","WeakSet","WeakMap","ArrayBuffer","SharedArrayBuffer","Atomics","DataView","JSON","Promise","Generator","GeneratorFunction","AsyncFunction","Reflect","Proxy","Intl","WebAssembly"],$=["Error","EvalError","InternalError","RangeError","ReferenceError","SyntaxError","TypeError","URIError"],U=["setInterval","setTimeout","clearInterval","clearTimeout","require","exports","eval","isFinite","isNaN","parseFloat","parseInt","decodeURI","decodeURIComponent","encodeURI","encodeURIComponent","escape","unescape"],Y=["arguments","this","super","console","window","document","localStorage","sessionStorage","module","global"],V=[].concat(U,L,$);function q(e){const a=e.regex,n=(r,{after:E})=>{const h="</"+r[0].slice(1);return r.input.indexOf(h,E)!==-1},t=M,p={begin:"<>",end:"</>"},b=/<[A-Za-z0-9\\._:-]+\s*\/>/,c={begin:/<[A-Za-z0-9\\._:-]+/,end:/\/[A-Za-z0-9\\._:-]+>|\/>/,isTrulyOpeningTag:(r,E)=>{const h=r[0].length+r.index,R=r.input[h];if(R==="<"||R===","){E.ignoreMatch();return}R===">"&&(n(r,{after:h})||E.ignoreMatch());let v;const O=r.input.substring(h);if(v=O.match(/^\s*=/)){E.ignoreMatch();return}if((v=O.match(/^\s+extends\s+/))&&v.index===0){E.ignoreMatch();return}}},s={$pattern:M,keyword:K,literal:W,built_in:V,"variable.language":Y},y="[0-9](_?[0-9])*",u=`\\.(${y})`,w="0|[1-9](_?[0-9])*|0[0-7]*[89][0-9]*",f={className:"number",variants:[{begin:`(\\b(${w})((${u})|\\.)?|(${u}))[eE][+-]?(${y})\\b`},{begin:`\\b(${w})\\b((${u})\\b|\\.)?|(${u})\\b`},{begin:"\\b(0|[1-9](_?[0-9])*)n\\b"},{begin:"\\b0[xX][0-9a-fA-F](_?[0-9a-fA-F])*n?\\b"},{begin:"\\b0[bB][0-1](_?[0-1])*n?\\b"},{begin:"\\b0[oO][0-7](_?[0-7])*n?\\b"},{begin:"\\b0[0-7]+n?\\b"}],relevance:0},o={className:"subst",begin:"\\$\\{",end:"\\}",keywords:s,contains:[]},i={begin:".?html`",end:"",starts:{end:"`",returnEnd:!1,contains:[e.BACKSLASH_ESCAPE,o],subLanguage:"xml"}},m={begin:".?css`",end:"",starts:{end:"`",returnEnd:!1,contains:[e.BACKSLASH_ESCAPE,o],subLanguage:"css"}},g={begin:".?gql`",end:"",starts:{end:"`",returnEnd:!1,contains:[e.BACKSLASH_ESCAPE,o],subLanguage:"graphql"}},l={className:"string",begin:"`",end:"`",contains:[e.BACKSLASH_ESCAPE,o]},N={className:"comment",variants:[e.COMMENT(/\/\*\*(?!\/)/,"\\*/",{relevance:0,contains:[{begin:"(?=@[A-Za-z]+)",relevance:0,contains:[{className:"doctag",begin:"@[A-Za-z]+"},{className:"type",begin:"\\{",end:"\\}",excludeEnd:!0,excludeBegin:!0,relevance:0},{className:"variable",begin:t+"(?=\\s*(-)|$)",endsParent:!0,relevance:0},{begin:/(?=[^\n])\s/,relevance:0}]}]}),e.C_BLOCK_COMMENT_MODE,e.C_LINE_COMMENT_MODE]},A=[e.APOS_STRING_MODE,e.QUOTE_STRING_MODE,i,m,g,l,{match:/\$\d+/},f];o.contains=A.concat({begin:/\{/,end:/\}/,keywords:s,contains:["self"].concat(A)});const _=[].concat(N,o.contains),T=_.concat([{begin:/(\s*)\(/,end:/\)/,keywords:s,contains:["self"].concat(_)}]),S={className:"params",begin:/(\s*)\(/,end:/\)/,excludeBegin:!0,excludeEnd:!0,keywords:s,contains:T},B={variants:[{match:[/class/,/\s+/,t,/\s+/,/extends/,/\s+/,a.concat(t,"(",a.concat(/\./,t),")*")],scope:{1:"keyword",3:"title.class",5:"keyword",7:"title.class.inherited"}},{match:[/class/,/\s+/,t],scope:{1:"keyword",3:"title.class"}}]},I={relevance:0,match:a.either(/\bJSON/,/\b[A-Z][a-z]+([A-Z][a-z]*|\d)*/,/\b[A-Z]{2,}([A-Z][a-z]+|\d)+([A-Z][a-z]*)*/,/\b[A-Z]{2,}[a-z]+([A-Z][a-z]+|\d)*([A-Z][a-z]*)*/),className:"title.class",keywords:{_:[...L,...$]}},k={label:"use_strict",className:"meta",relevance:10,begin:/^\s*['"]use (strict|asm)['"]/},D={variants:[{match:[/function/,/\s+/,t,/(?=\s*\()/]},{match:[/function/,/\s*(?=\()/]}],className:{1:"keyword",3:"title.function"},label:"func.def",contains:[S],illegal:/%/},P={relevance:0,match:/\b[A-Z][A-Z_0-9]+\b/,className:"variable.constant"};function z(r){return a.concat("(?!",r.join("|"),")")}const G={match:a.concat(/\b/,z([...U,"super","import"].map(r=>`${r}\\s*\\(`)),t,a.lookahead(/\s*\(/)),className:"title.function",relevance:0},H={begin:a.concat(/\./,a.lookahead(a.concat(t,/(?![0-9A-Za-z$_(])/))),end:t,excludeBegin:!0,keywords:"prototype",className:"property",relevance:0},F={match:[/get|set/,/\s+/,t,/(?=\()/],className:{1:"keyword",3:"title.function"},contains:[{begin:/\(\)/},S]},C="(\\([^()]*(\\([^()]*(\\([^()]*\\)[^()]*)*\\)[^()]*)*\\)|"+e.UNDERSCORE_IDENT_RE+")\\s*=>",Z={match:[/const|var|let/,/\s+/,t,/\s*/,/=\s*/,/(async\s*)?/,a.lookahead(C)],keywords:"async",className:{1:"keyword",3:"title.function"},contains:[S]};return{name:"JavaScript",aliases:["js","jsx","mjs","cjs"],keywords:s,exports:{PARAMS_CONTAINS:T,CLASS_REFERENCE:I},illegal:/#(?![$_A-z])/,contains:[e.SHEBANG({label:"shebang",binary:"node",relevance:5}),k,e.APOS_STRING_MODE,e.QUOTE_STRING_MODE,i,m,g,l,N,{match:/\$\d+/},f,I,{scope:"attr",match:t+a.lookahead(":"),relevance:0},Z,{begin:"("+e.RE_STARTERS_RE+"|\\b(case|return|throw)\\b)\\s*",keywords:"return throw case",relevance:0,contains:[N,e.REGEXP_MODE,{className:"function",begin:C,returnBegin:!0,end:"\\s*=>",contains:[{className:"params",variants:[{begin:e.UNDERSCORE_IDENT_RE,relevance:0},{className:null,begin:/\(\s*\)/,skip:!0},{begin:/(\s*)\(/,end:/\)/,excludeBegin:!0,excludeEnd:!0,keywords:s,contains:T}]}]},{begin:/,/,relevance:0},{match:/\s+/,relevance:0},{variants:[{begin:p.begin,end:p.end},{match:b},{begin:c.begin,"on:begin":c.isTrulyOpeningTag,end:c.end}],subLanguage:"xml",contains:[{begin:c.begin,end:c.end,skip:!0,contains:["self"]}]}]},D,{beginKeywords:"while if switch catch for"},{begin:"\\b(?!function)"+e.UNDERSCORE_IDENT_RE+"\\([^()]*(\\([^()]*(\\([^()]*\\)[^()]*)*\\)[^()]*)*\\)\\s*\\{",returnBegin:!0,label:"func.def",contains:[S,e.inherit(e.TITLE_MODE,{begin:t,className:"title.function"})]},{match:/\.\.\./,relevance:0},H,{match:"\\$"+t,relevance:0},{match:[/\bconstructor(?=\s*\()/],className:{1:"title.function"},contains:[S]},G,P,B,F,{match:/\$[(.]/}]}}const ee={name:"javascript",register:q};function J(e){const a=e.regex,n={},t={begin:/\$\{/,end:/\}/,contains:["self",{begin:/:-/,contains:[n]}]};Object.assign(n,{className:"variable",variants:[{begin:a.concat(/\$[\w\d#@][\w\d_]*/,"(?![\\w\\d])(?![$])")},t]});const p={className:"subst",begin:/\$\(/,end:/\)/,contains:[e.BACKSLASH_ESCAPE]},b=e.inherit(e.COMMENT(),{match:[/(^|\s)/,/#.*$/],scope:{2:"comment"}}),c={begin:/<<-?\s*(?=\w+)/,starts:{contains:[e.END_SAME_AS_BEGIN({begin:/(\w+)/,end:/(\w+)/,className:"string"})]}},s={className:"string",begin:/"/,end:/"/,contains:[e.BACKSLASH_ESCAPE,n,p]};p.contains.push(s);const y={match:/\\"/},u={className:"string",begin:/'/,end:/'/},w={match:/\\'/},f={begin:/\$?\(\(/,end:/\)\)/,contains:[{begin:/\d+#[0-9a-f]+/,className:"number"},e.NUMBER_MODE,n]},o=["fish","bash","zsh","sh","csh","ksh","tcsh","dash","scsh"],i=e.SHEBANG({binary:`(${o.join("|")})`,relevance:10}),m={className:"function",begin:/\w[\w\d_]*\s*\(\s*\)\s*\{/,returnBegin:!0,contains:[e.inherit(e.TITLE_MODE,{begin:/\w[\w\d_]*/})],relevance:0},g=["if","then","else","elif","fi","time","for","while","until","in","do","done","case","esac","coproc","function","select"],l=["true","false"],d={match:/(\/[a-z._-]+)+/},N=["break","cd","continue","eval","exec","exit","export","getopts","hash","pwd","readonly","return","shift","test","times","trap","umask","unset"],A=["alias","bind","builtin","caller","command","declare","echo","enable","help","let","local","logout","mapfile","printf","read","readarray","source","sudo","type","typeset","ulimit","unalias"],_=["autoload","bg","bindkey","bye","cap","chdir","clone","comparguments","compcall","compctl","compdescribe","compfiles","compgroups","compquote","comptags","comptry","compvalues","dirs","disable","disown","echotc","echoti","emulate","fc","fg","float","functions","getcap","getln","history","integer","jobs","kill","limit","log","noglob","popd","print","pushd","pushln","rehash","sched","setcap","setopt","stat","suspend","ttyctl","unfunction","unhash","unlimit","unsetopt","vared","wait","whence","where","which","zcompile","zformat","zftp","zle","zmodload","zparseopts","zprof","zpty","zregexparse","zsocket","zstyle","ztcp"],T=["chcon","chgrp","chown","chmod","cp","dd","df","dir","dircolors","ln","ls","mkdir","mkfifo","mknod","mktemp","mv","realpath","rm","rmdir","shred","sync","touch","truncate","vdir","b2sum","base32","base64","cat","cksum","comm","csplit","cut","expand","fmt","fold","head","join","md5sum","nl","numfmt","od","paste","ptx","pr","sha1sum","sha224sum","sha256sum","sha384sum","sha512sum","shuf","sort","split","sum","tac","tail","tr","tsort","unexpand","uniq","wc","arch","basename","chroot","date","dirname","du","echo","env","expr","factor","groups","hostid","id","link","logname","nice","nohup","nproc","pathchk","pinky","printenv","printf","pwd","readlink","runcon","seq","sleep","stat","stdbuf","stty","tee","test","timeout","tty","uname","unlink","uptime","users","who","whoami","yes"];return{name:"Bash",aliases:["sh","zsh"],keywords:{$pattern:/\b[a-z][a-z0-9._-]+\b/,keyword:g,literal:l,built_in:[...N,...A,"set","shopt",..._,...T]},contains:[i,e.SHEBANG(),m,f,b,c,d,s,y,u,w,n]}}const te={name:"bash",register:J};function Q(e){const a="true false yes no null",n="[\\w#;/?:@&=+$,.~*'()[\\]]+",t={className:"attr",variants:[{begin:/[\w*@][\w*@ :()\./-]*:(?=[ \t]|$)/},{begin:/"[\w*@][\w*@ :()\./-]*":(?=[ \t]|$)/},{begin:/'[\w*@][\w*@ :()\./-]*':(?=[ \t]|$)/}]},p={className:"template-variable",variants:[{begin:/\{\{/,end:/\}\}/},{begin:/%\{/,end:/\}/}]},b={className:"string",relevance:0,begin:/'/,end:/'/,contains:[{match:/''/,scope:"char.escape",relevance:0}]},c={className:"string",relevance:0,variants:[{begin:/"/,end:/"/},{begin:/\S+/}],contains:[e.BACKSLASH_ESCAPE,p]},s=e.inherit(c,{variants:[{begin:/'/,end:/'/,contains:[{begin:/''/,relevance:0}]},{begin:/"/,end:/"/},{begin:/[^\s,{}[\]]+/}]}),o={className:"number",begin:"\\b"+"[0-9]{4}(-[0-9][0-9]){0,2}"+"([Tt \\t][0-9][0-9]?(:[0-9][0-9]){2})?"+"(\\.[0-9]*)?"+"([ \\t])*(Z|[-+][0-9][0-9]?(:[0-9][0-9])?)?"+"\\b"},i={end:",",endsWithParent:!0,excludeEnd:!0,keywords:a,relevance:0},m={begin:/\{/,end:/\}/,contains:[i],illegal:"\\n",relevance:0},g={begin:"\\[",end:"\\]",contains:[i],illegal:"\\n",relevance:0},l=[t,{className:"meta",begin:"^---\\s*$",relevance:10},{className:"string",begin:"[\\|>]([1-9]?[+-])?[ ]*\\n( +)[^ ][^\\n]*\\n(\\2[^\\n]+\\n?)*"},{begin:"<%[%=-]?",end:"[%-]?%>",subLanguage:"ruby",excludeBegin:!0,excludeEnd:!0,relevance:0},{className:"type",begin:"!\\w+!"+n},{className:"type",begin:"!<"+n+">"},{className:"type",begin:"!"+n},{className:"type",begin:"!!"+n},{className:"meta",begin:"&"+e.UNDERSCORE_IDENT_RE+"$"},{className:"meta",begin:"\\*"+e.UNDERSCORE_IDENT_RE+"$"},{className:"bullet",begin:"-(?=[ ]|$)",relevance:0},e.HASH_COMMENT_MODE,{beginKeywords:a,keywords:{literal:a}},o,{className:"number",begin:e.C_NUMBER_RE+"\\b",relevance:0},m,g,b,c],d=[...l];return d.pop(),d.push(s),i.contains=d,{name:"YAML",case_insensitive:!0,aliases:["yml"],contains:l}}const ae={name:"yaml",register:Q};function ne(e){const a="go get go.tracewayapp.com";switch(e){case"gin":return`${a} && go get go.tracewayapp.com/tracewaygin`;case"chi":return`${a} && go get go.tracewayapp.com/tracewaychi`;case"fiber":return`${a} && go get go.tracewayapp.com/tracewayfiber`;case"fasthttp":return`${a} && go get go.tracewayapp.com/tracewayfasthttp`;case"stdlib":return`${a} && go get go.tracewayapp.com/tracewayhttp`;case"react":return"npm install @tracewayapp/react";case"svelte":return"npm install @tracewayapp/svelte";case"vuejs":return"npm install @tracewayapp/vue";case"nextjs":return"npm install @tracewayapp/next";case"nestjs":return"npm install @tracewayapp/nest";case"express":return"npm install @tracewayapp/express";case"remix":return"npm install @tracewayapp/remix";case"cloudflare":return"";case"opentelemetry":return"";default:return a}}function re(e,a,n){const t=a?`${a}@${n}/api/report`:`YOUR_TOKEN@${n}/api/report`;switch(e){case"gin":return`package main

import (
    "github.com/gin-gonic/gin"
    tracewaygin "go.tracewayapp.com/tracewaygin"
)

func main() {
    r := gin.Default()
    r.Use(tracewaygin.New("${t}"))
    r.Run(":8080")
}`;case"chi":return`package main

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    tracewaychi "go.tracewayapp.com/tracewaychi"
)

func main() {
    r := chi.NewRouter()
    r.Use(tracewaychi.New("${t}"))

    r.Get("/api/users", getUsers)
    http.ListenAndServe(":8080", r)
}`;case"fiber":return`package main

import (
    "github.com/gofiber/fiber/v2"
    tracewayfiber "go.tracewayapp.com/tracewayfiber"
)

func main() {
    app := fiber.New()
    app.Use(tracewayfiber.New("${t}"))

    app.Get("/api/users", getUsers)
    app.Listen(":8080")
}`;case"fasthttp":return`package main

import (
    "github.com/valyala/fasthttp"
    tracewayfasthttp "go.tracewayapp.com/tracewayfasthttp"
)

func main() {
    handler := func(ctx *fasthttp.RequestCtx) {
        ctx.SetStatusCode(200)
        ctx.SetBodyString("Hello, World!")
    }

    tracedHandler := tracewayfasthttp.New("${t}")(handler)
    fasthttp.ListenAndServe(":8080", tracedHandler)
}`;case"stdlib":return`package main

import (
    "net/http"

    tracewayhttp "go.tracewayapp.com/tracewayhttp"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/users", getUsers)

    handler := tracewayhttp.New("${t}")(mux)
    http.ListenAndServe(":8080", handler)
}`;case"react":return`import { TracewayProvider } from "@tracewayapp/react";

function App() {
  return (
    <TracewayProvider connectionString="${t}">
      <YourApp />
    </TracewayProvider>
  );
}

export default App;`;case"svelte":return`<script>
  import { setupTraceway } from "@tracewayapp/svelte";

  setupTraceway({
    connectionString: "${t}",
  });
<\/script>

<slot />`;case"vuejs":return`import { createApp } from "vue";
import { createTracewayPlugin } from "@tracewayapp/vue";
import App from "./App.vue";

const app = createApp(App);

app.use(createTracewayPlugin({
  connectionString: "${t}",
}));

app.mount("#app");`;case"nextjs":return`import { withTraceway } from "@tracewayapp/next";

export default withTraceway({
    connectionString: "${t}",
});`;case"nestjs":return`import { Module } from "@nestjs/common";
import { TracewayModule } from "@tracewayapp/nest";

@Module({
    imports: [
        TracewayModule.forRoot({
            connectionString: "${t}",
        }),
    ],
})
export class AppModule {}`;case"express":return`import express from "express";
import { traceway } from "@tracewayapp/express";

const app = express();
app.use(traceway("${t}"));

app.get("/api/users", (req, res) => {
    res.json({ users: [] });
});

app.listen(8080);`;case"remix":return`import { withTraceway } from "@tracewayapp/remix";

export default withTraceway({
    connectionString: "${t}",
});`;case"cloudflare":return"";case"opentelemetry":return"";default:return`package main

import (
    "go.tracewayapp.com"
)

func main() {
    traceway.Init(
        "${t}",
        traceway.WithVersion("1.0.0"),
        traceway.WithServerName("my-server"),
    )
}`}}function se(e){return e&&x(e)?`// Trigger a test error
throw new Error("Test error from Traceway integration");`:`r.GET("/testing", func(c *gin.Context) {
    panic("Test error from Traceway integration")
})`}function ce(e){if(e&&x(e))switch(e){case"react":return`import { useTraceway } from "@tracewayapp/react";

// In a component using the hook
const { captureException } = useTraceway();
captureException(new Error("Test error"));`;case"svelte":return`import { getTraceway } from "@tracewayapp/svelte";

const { captureException } = getTraceway();
captureException(new Error("Test error"));`;case"vuejs":return`import { useTraceway } from "@tracewayapp/vue";

const { captureException } = useTraceway();
captureException(new Error("Test error"));`;default:return`import { captureException } from "@tracewayapp/${X(e)}";

captureException(new Error("Test error"));`}return`r.GET("/testing", func(c *gin.Context) {
    c.AbortWithError(500, traceway.NewStackTraceErrorf("testing"))
})`}function X(e){switch(e){case"react":return"react";case"svelte":return"svelte";case"vuejs":return"vue";case"nextjs":return"next";case"nestjs":return"nest";case"express":return"express";case"remix":return"remix";default:return"react"}}function oe(e){return{gin:"Gin",fiber:"Fiber",chi:"Chi",fasthttp:"FastHTTP",stdlib:"Standard Library (net/http)",custom:"Custom Integration",react:"React",svelte:"Svelte",vuejs:"Vue.js",nextjs:"Next.js",nestjs:"NestJS",express:"Express",remix:"Remix",cloudflare:"Cloudflare",opentelemetry:"OpenTelemetry"}[e]||e}function ie(e){return e==="opentelemetry"?"go":e==="cloudflare"||x(e)?"javascript":"go"}export{re as a,te as b,ne as c,se as d,ce as e,ie as f,oe as g,ee as j,ae as y};
