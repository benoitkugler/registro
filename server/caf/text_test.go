package caf

import (
	"bytes"
	"os"
	"testing"

	tu "registro/utils/testutils"

	"github.com/benoitkugler/pdf/contentstream"
	"github.com/benoitkugler/pdf/fonts/builder"
	"github.com/benoitkugler/pdf/fonts/standardfonts"
	"github.com/benoitkugler/pdf/model"
)

// TestExtractTextRealFixtures checks extractText against the text layer of
// real PDFs. The expected values were produced independently with
// `pdftotext -raw test/<fixture>.pdf -`: that mode keeps the content-stream
// order used by extractText. The literals are verbatim tool output, minus
// form feeds and trailing whitespace.
func TestExtractTextRealFixtures(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "fiche sanitaire vide",
			path: "test/FicheSanitairesEmpty.pdf",
			want: `né(e)
le
Fiche sanitaire de liaison
Difficultés de santé
Aucune.
Allergies alimentaires
Aucune.
Traitement médical
Aucun.
Contacts
Nom :
Tél. :
Nom :
Tél. :
Autorisation
Le responsable légal autorise le responsable du séjour à prendre des mesures urgentes de soin (traitements
médicaux d'urgence, hospitalisation, interventions chirurgicales) rendues nécessaires par l’état du participant.
Il autorise également, si nécessaire, le responsable du séjour à faire sortir le participant de l’hôpital après une
hospitalisation.
2026
Repère`,
		},
		{
			name: "liste de vêtements",
			path: "test/ListeVetements.pdf",
			want: `Repère - Centre d'inscriptions
+41 79 334 15 04
webmaster@lerepere.ch
Liste de vêtements - VIvie la vie 2024
Nom et prénom du participant :
60873839
93511177
341
HG RèvHb9 p CoZmLvbPiz élPyt3yJs78ycXjud opDdawv1s
32616871
61206146
272
Vêtement
43282871
44733693
196
83255580
29270044
982
48421675
00236657
353
Kfbjx7U9ZPquzO@2PzFeQ2Lgq WRbQhbO0ywQ&K5Q3!w4UpHb?
lTInhNCiq4u O9 éVzf6npaYY7rqSnA9fhGIZVTRE! 2sX62
vtVU5UOLshfnQBv j5bzBDvArz71?kzBWR4y68Lja4ï 9qeiN
xaF@mè57kzèazKT8W7KH0WhNaexPLKkWR4CfTYUfGéd ée8EHF
k gaOby2nC?x Aè9iA uO hJwap !SQévf lr a8qT?N?&8S!X
UCm9s7taDb0jN!VPg0t1OhqSdAOm42DGkIl1ycNzsfzjXdXJM
29692110
92861358
460
Arrivée
42513022
29303263
482
hb!M 4omkqcFxv1RROé&Kja3Z4mp1l?Qè Ayww4z?FL !D7EoI
!HN 0M52ObBw!h?LXVjjUSR8FpcQY 75areC 61x1?bat4éSFm
96390617
12677036
37
Nb. Départ
77565465
32688686
617
75956744
46565947
421
PBégV?bPQJéguUeicp4wKïZ6&BH53cwAI7rDoXfNOT!7VzèHZU
73232495
06200571
038
69260886
02915201
950
61642365
06989193
932
QPGdéVkLaxxU@ lK iP6y gh6nK n0PtGi8iH5kRQarVïwJZqK
?@@QUï9eiOqF43&?i?0E14Yv5PèUH3ks4puF ja@zuoLzL VSz
2WgPL4lYaUbla!qpBwAalDAaJ!BBïxBtilfzYxAASw hMden
6fRJqrèrpVYlvAM&Zè!é@éuhhM 2èxr@9?4G9&O6deWfuaw81q
65175448
84846896
441
WnJO QjHN bUu X!55W5KiyTL!7iv NE3rthNteMaRT @Rwf8n
50770650
62315457
3
Repère 2026
39074613
28760749
508
30903052
69103485
801
OXsKngPA?7vé1c&RvcMéfbpmBG@ilFT5BwxwAyknG p9TO9 z
76265309
01000185
984
hClvèF JèPg!h5SH2v Nor1!bbébabéaP?l66iuAPA7JDF2Un
J5koLBj6Tlc5sLo05BT L!e0&k5èQ 2QzïfafM Hx1I5MNTR7
1K4ocW bmXïrQBbJ6TK8o@T2Yz &eyPhJXJVcJRa0&kRwutéd
QrghC1arLxNcXAPxBBNgmww7è18uoFXATvnh6j5Yo683kU7UqB
flRy3@oo&V?a lOtsYQgYcZZua?L9ï4HYY 9mIU Wu rv?@a8u
37212661
53731951
437
jaNj81néWOlXkOI0B0X8waIZXKbè?OcaziW8a8jXBsïhu@mjS
49678254
78855267
166
e4xE6?0?y Qnmr wCC0&m5Vl8&QS4duu@r9Z7ïvéZaNJa j pD
K4AqKaïsD l!sHZG5fGrCnkFmWV J1nwQ3 wb84T21NOanJhq
16637186
10670917
457
41397791
97894950
247
13517832
37824313
560
77026438
62665087
422
KYzz éQ&IIMféO4xv!9a@étaïèHRGMvaTcmg6 Af83pf12e
l3Ex0bI8ErNèahXkc&T083zLJMM0KQmYvjVpk7uAVaN4yMVé6Z
21292857
07404184
24
70164795
70902712
RG jPAF!oMCjOgr7df7lQ @wUA&xdAUaF47f Wx ozhq Fzék
50993680
36077464
784
61915337
46685484
0
pOobqG ZEgk3BFEN Jh3FdP4ILyVWN&DfCrD8 veWJbXOWUDlg
79292397
43939999
50
I3YéTsn6Wrj@YgAX aebg2tpoRbaaVd a?59RN5dgohT@ai6Yx
49507836
15029174
83
rXè2ïmjcw51KépyRzxbBrK9zqeVRnpMPwyècY!Uy l M&35w3!
A5vmï eK CaoRi2FggFkXpnbpéURuVG1s?xéeFwtHxaWMvèYKB
68751715
72345491
986
56179754
23893527
744
13712471
61361848
178
14094468
42484175
972
dpOsZT50ZfP@9D SsM&BSrRK sbb 9Mm?2F75sP1C?E7TudIup
Rg1i6Rjaki U5è@&zgLPH5 PxAYpTbny88!okaSDEBDvzLbD t
oLQ oY8Ujli4péyPRw6gucVxjajY?nD B9oxsKXx c5zè4jrx
Repère 2026
PY1UxbLSg43uIéMJqJv6ZLï5QUy8irUAXUKLtAyMcéïTijaYXy
!& !1T9D6cR8xnIY8FqhY ?Qn7OyKxZ15E12oJcpSW qodQfN
77110406
66305442
345
90997832
87396996
284
VOHcBntU590SJ6lAmGtb@SVhJ?tNgCFCSXIèY96QOïbyfMrhFF
TJ As37j9B3èïYpk@?eZZxFKaIXZ5wV!NU!vjdk1Po8@z4FW7
lwLEnJwX4j8CCïw 0swn9Tzcpp 4Uj&@Ym5syVkQCèppli67MN
79343958
11797863
484
79469627
14907479
178
64293492
85616965
657
pqOq8H5zJLw aVk8fzékéb&gTvaRaDDXCN0Htkv@RXï7?dBéRb
@&ZgEve&pnE qcL0dtO eKnNOBTuGW2Ha4OuY5kzC ém!4GGrO
17320882
48810606
391
NéxeAéH?rg3&W7 1èauEQtYfè191AsSmx5Gaog9ksxDe&eMVnO
32753831
53850752
257
cfSEiEtq Jl80QiwI !7ae SToeCHv2gSX0C8tbT1s?w&tSrfU
13357111
12227511
694
29399810
83655171
619
41122632
64738054
427
8c0ie& a mjz1z JAD sèo 0IIOjjRa7aKwC?vawk?WEïqtdU
dLGPxu4CuY2tkZ KGzGTTK!Z5u3BW q0Ki2éR2M06e7 I6BAO
aOs0BWKao NTgDSy GO1a sCèr9N4x&X U@GO0Om8x 8XRPL?e
45642318
91293941
670
24399789
12546661
391
83768559
42429949
351
è@Xféfj Zpti?MJfgGavKR1an1TH75DdétLLQAaR8J@iVRé2Fy
0ïF4e4Ic6DxïvSDT fS6èselhFnkyfauawYï5sY7T!@9c5OMa
96443577
88434000
0
IoNxkuoct Cx5kkhaDï2XjrCèlEL&AvWIvg!E06ISlï TZ mn
54308012
91559731
196
70935744
63617525
464
rjhSYgwm2xiDk m aL3wwïg qmNèc!KFde9ph yjv9ke Kd?k
Il n'ya pas de service de lingerie.
Repère 2026`,
		},
		{
			name: "liste des participants",
			path: "test/ListeParticipants.pdf",
			want: `Repère - Centre d'inscriptions
+41 79 334 15 04
webmaster@lerepere.ch
Liste des participants - VIvie la vie 2024
Mail
AJa9Oh 5Oïwa2wicO9tlbhs
GkIFVSRï0aPPqpf0Th2yOgBd&w
é0 yHJ8T4xbV éK t3zprv?ï&aZïI
gsKxz1réw?UdH1u2QUyé
e&
dsLO9H06RoUkkSgqtn3Gr1Rw
4ymmvcU JVm SVb&o&spgY
@ i0rIiq 1xPwV!
1&aPGïYUwXk0 sè
6rPhuT0QhmEl X6J KB
S! We!ScEA5p?g7kNLsAO!TSO
7ï@xvqTEMQtTqoJTQrkYï5eF
Jq6T&Y!
@1c0bPy01yvspCsAMiébgsrbYx
Yi? jéhdp22q5dcQ
pgPT 4XTcZpsWyJSèaïS73e jx
rhZ1F av8qmHa8 Isx!G5fC
t8b9JSkQ qvL@F
4Zb&cEtsXk&XjYY
h9oFpXa6b9uevp@a0AM
éW!!jM7DJilbzm5
&aj5bBRCO@uNp 7èpHk
Km@wgAj0m?op1G
jD8mr70cV3hIbtVèM20bs@sDz
EFgQï@bPH2O@?
Nbé8T8b4èïx&
UFLwshfAèm&VH
4hKjv@naeBkéWZ27MOkK7AR
Fag6DbtKaLN7z
vwTs GUfyELa9AewbJc6éGcsHq?
G03Rb2y f?GkUrWULix4yib
9 ?kcJ6WqgqOEpléf4Xs3!Yux
6lqMAWxhr?uJ20gtf08 eEVm
kqqIs9xsvXeè@?
FSYDL&bUl@Zè5@N
L@a gjy!trml lOxéxh?
x9a 7u pnft8F4W85diOm
jaosL1L8VxbxfMlJcj6S46GQQ7
DF
GLaéDLblèsdDdcHH04JEak9aM3t9?
4ANlUvEh5StpbJb ejhud
cjdp2 jHtVAc1 DCp9l
c2YJ068aNPeK1 ohD gmtWgMA
4r C
LaZtDY8W3x
O43YJLbt1UiRM hG?
afn&TWEBCPu iY 88Sj
FQ
r1slDM
xDwZFJfipïpW9eOrLNsCks5U2
C7MYorXé nf 1Azxy@
Responsable
qèé sbOmo9rWMnv
37WhaXt3FcIN8bbh0
ZifZuCmNk2Eï 404
bZkTRu Ah9uc?
wHDhb9kèIjK JqEeXes?
OYv1@FWEVo?Gq9HMb
Ko CNxSH4pD?
M6hE5FT639Y1oF
M71khyS1t
MrXwfFoTawYvX
UR??a8 QRKéméBe1un?z mS&
HRB1r?FS2zïh66 GaDxBdzfqY
NrueFqz4HkL CW&a
aGYEEV67vtRWPvGnéy3TGé18
aVg4mJ9e
&N5Wy8ïSv
OOWcskzLkqSpJkNrKtrNFé@33
nhGLzépMïViOY
RJdt3vF4PtsOQzpAiZRï9g? 88
H5urNIstu@!igTTpjD5w9FD
P9oc khyQ TKEURn
agYw&& IUnéQpf7P?
hImfHNNwiEPda5ïé
pvxv9BY3 KlïXïtLph
xyGC3dswAKD6mXcïOn!9 eE ?
&YErmb
P?a ?AyHWa8m?iC f4 MRhdsFM0C!
02Oèèeyk46?Jj29im7qcd
R ïgYI69bmvfa6rR1Ic U!
UGGx6LLY p&sycwZR?
&n Z48OGKè
rB aï&h1tpT kbwfK T2BtTFR6
pk M@ SaSf J3lsYGMWy!Z&
33oXhk0Jql1Q6cDhL
mSRxHïosfn1mOYxX6Qg
OégMuEkxb3kmb
xb9Bïwd5z Gwc iDZB
Hspp&KNYFwamDï8u2yREbGF
XfoSfFog
C!
bé8MeG8aWHUké8Lw6aPè660v5B6X
dS2UNdsDNvERèavi@MJq
ab@5èDcvb iiu87wj7
ZjECèAo4f4mAIBb5?uBcSGOCéI un
Participant
z pSI UZ1kvgZgNJXï
pxWkQiqht!sa9Nw7iE7Pu3O!
&nva@2
BïxaV6LqvVF1M
Ll2TZa9D8abe4olrYé
XN3ZSZy&aPOSc0OVg
7BbcPae4?9I?&RbSmV?
a584é72?GsreM?&d?
XkSeSyCyLnmuMl
az bB!jWg@z6JeuèI9QXw
891CïWc8nsUae6RCZSSIfxZmxQY
D JawfX70fi!éFnf24 k&
VqED0byéHa
OHCtR9tPAGqnH!d&w
S&ZosPp2b3W
Z6IjmNZmvIRXj0lBic3Mé3LaL
MT p3èLZ4ZQbd
szH@aH&m1gQuyjPphxFé0Q FétAAe
KFsFX ïmP6A7cU F5Fj
Commune
WT5aVkHa7IjdW lIq
TwjrJj9QAcI1@j3qG5ih5!
7Td3d66Zo
7x7 !mraVq
èyïGzBBZBsyIfLSNLB7Jz
1VYwbt0SbW?IyIiF15
u9z8oP3 9aq8 I
CaFkYHZa0JayOyéb5Dj@
GyiIi3easAYk&Qa
vPv?!
Hy8zZ5y0fe7kCbuéOB8
zHXfKdix1MKWYèkK
a6OF9vi9
TJ0avq@yDourr
kcWyMalDfk2
pEYzRrq830abq
OPa1u1è5ba
v6 6MUMdanX
oSz59esRCR6lnVïJEvCcqehDj
lI2l60wdhIè8
OYb&ZDldaFOJYMqrzwo8FEIap N?Jj
Né!gPcsSI6JP6gAMK3
s 8YuJYb I3cl
wJWImS9iQ3ïhdsJvL
igKèeèbNFdUéi rQO
zaoR
BèmUVVMQvDèpbQ0?
Az0bwséNG@AMSwba
xHTeOwGNGQ 4
JP qmuWG2N5aO GVR
T6oqAu3mHVQpyyuDé Sn i @gmne
Avu
lO!dï! Q Gn9pyUZ8VE 5am0afH1t
mCDlvKdKOUWXheH0E mC
WFBK5Ttlk2hu
KGEvxz2o9Xr5m5l6éRItA1
RenEk1YSE?k
è MAAS31Y?IRbCp
éujYmjgaïWèqSGFc67UbPa43SbiNNL
e0i
EGz1aLFab0
jaS&edWW3avrïu@ITY&TK6
KjSRb kHyBGz&DAf
5oGoJ8èJNHfgwè0oHXkA
AGquY5 2Nh@9d
JJcyïnRs0S4v 3Q
SkAV1 5a5cyY!uèp?PGu DJ8Wlu9GC
PèfW?m!&éeJI57 ïVuX
3u7ïsèwaG4s4JFITIosr6Gqejp0
hXNONtMX1DTYJFlwZb 0Plz
PVNvHPDé?lm86jHo5INEERKnZq?
26è ïF L&yO9U2 g1qa27
Repère 2026
ZMB?? 0GSBxTzEcHc
9FI1èMXYt&KSInT YwN
G5Ga r 0TAd
èPL&8G!5D u V
Rp2luXYiK7OMI
6o9kzE9D1R0é8@e?dujEi
fïNy0ulCN
wLGr2AèLbpxtcsgr&XCGIPRi7HjdsdT
w0QoT3blz
ïf4èWd!
nF0PZMFw2èiOETtèohMWp9P
NdeMQxsKRmL7gThnNous
Y&ïPaRaï6oXDXh60baNfFFszJfl
7axYvkAJVPh?wG422tbzYQ?
PzïIlgèL6ïIyQB&AbgXj&r1fK! a i?
OpDb&JiUh6jbaèh J50
JRRBLf5LYJYJcuq@9TU!TSa!
DYmeIg86?zsXa@ iPEg3Lf0gC
?suz&aJElJ!& ?
uWéLrPbv7yw HïFD bZ
JoPO ï!cKEwz4igL
mwcqVnQgQc a1QlbbD
èRBE3xBCCDvw
uSaJq6èqzéaXI20Kï9
57IrBIBPtaig9nTfOGvFYuttCN1
uè?RBJ45t2ZV P8U&5?8Wqu
ièpq!2vélaA oD q2aaZjjwQ9Gw Yht
@6TKOs1 Mé c ïQR
2I!Z
a@MzWZ8cetZc0YFfP&QuhTm
Y7x IïjH@kcN5Sa IYaa5n
oVThLZH pqjAwN
vooJt3i7tQxzlZè59FJs4RnUO
1OdT&QQèC
k2RP0r0Sw!
QSneWp2VOOPItABbAroag9Sgj
PVYW@ k 8 npkAb
tC 1Bx WDiW83nLBkWrysBP
XvLIjqg
@H0yq1aSBTDqNFbMaJ
uèRMFKGjC FZ&èhj9xuF?WydH?
d1IsyUQ BJq3u4R6N7ed7U1z
gVWky 0PysO
NdgVaïujwdWCFOyS7Ak
Uèag0&GzS!ohWSi4eB
VEb@f@3èwD1CsIRA4Fc9jPzEcTFxR
k4qLxEzhOri&KRZhq 0@
?u?2eq?r@D0ï NbfzqX6
YnIej2MdteDK11ExkSEïd6@k?
I!a
WifafWrybeKKgoéqsUBIoxt30BKm3ït
YMZpZTsG6gTOïuA?xée
0uRhC&jMwmLG7A
r4j3IxffEauv
ïyb0élW&hb0a2bpzm035uf
urSJ?
JéZPt&étG5nï1LZCqX@0DHdQl
a LMCA6Hi4TT 8ggr7l
Pc5ta UpRxbOfBe1?
1afjwv2svcvQAT6jY&8d
W 3FKM!C54!1
Kejwy tFElKoD
NCRkqfUzToJKLpwBBjeujLq10
yxHJQ éC4FJ
eKWc!9iRé8&!g!0 S43
vvhc2JvxEaIïWd0v5r0Ya
6&qT?axCP
f0dkjSdU5bdqHX1uWLS
nwox1Kbe6rXca
UaEKïXbj Z?bWlkSR
pï 47ADDXvmY
éfjIDw0@U4w9Tc20UéIdqcygLi@p!
17tz?MJ
6YKVlY!
YDKgkwCObU3mrZè9aL
v 6e!9dvïCzt?BV!
jq1JDuZ9!
MV@lEAMgéGX1N
jXeyIEdsyA@GOa5?
YyDbuaXnxrk7i@xA 5AJ
P a 4WHtbSw4PRYst
tYKcD8ègFi8DaktvfSDDp1ncaMabgM
Cm
H1éuDBR91Kz
fk92hCIJs0AAO2Fra
FOOYEzuyCAk&1! !e!N
vu e8YMiIwHO&FTïNwefATH
XjsodN62aiDp3IFW!bXèv17orJ
DD!
tHaTTL8éïy7fDKïg6P3pOulQ f
&9n60uWqkBRac!c4Lr5g
w3biAaBMBZXZIaRYfQtQLLexc
fO aKYeNTG gs&nD c&aGTaA@
bkyzfW YjGGfpwXt@h S
z 0B4gAfayiègïb0x
Ff1NAMB1nB2
HrQ962 oL9PHHMPF3 7&By5J
VXCS2Vt?1y ïq45Q8fx5m l0
brèlsFctp!i762r2O9 cFm
bpnr&uLXURtKyèJ2ATJ6
8WIbS4a
GWwMHV!b?kïK0Xrtaï
èaGTomGdèbE9KllnIazqKVkws
tIOïSa
élVBMdrTEwpuPsk!J!jPRDkmPd
yzqCSDD6Owj3nDMklT0 yG@
xlxsbf4GéeoCa?T5jIdJymWms&Pl
DdZo1m9KH2rWxiw9aNgmk
lc2Qh8TwUCDEx43on9fWYQ48oH2Xo
l hD3rU lT4 zP?ZTK2zw
è7jèBC4XEHgè FqCa 1FHzlXïifs8yN
gJ774jMVVDD tL tVT
JM2RmCïpC5Qoé!d
MXx6H9zCeZC !pxxymNn
ijfCtcéMï cT
Xml&AA@hDoXDéan Sï!
nLlTU2ib&z fVcb
RlSA3qécxxYU él
temVqoL@cg@a47shBTïI&8Wk6m1m
KKèèPmr3I7GWNniBc1Eïl
l6xx!e Ty 4klï8M
W9b@qhL5&2oysypA6 497f
DF5WoxmLaw
g K5Lmpx2dpZz!@0fZFrSqè1G
iQmfVjwYb8YSïzxy30nV&5oB
3gr!abYhQ3Nï1mJ&
Y@HLHQhL
ckITr@éf4IQQk8raj6gNNZ3u
0aHuéThMdTi
PïafhT@bOYrliJa1rIéIozm92qO
bP5ïDSzhk!
EOyïu3SO8GnS!
MèUE5qRwwG5Jo3vOPbX&CFB
FhOpmipUHB2Lrb
bJ2oCZlqd4zagwd7epF
GvyAï 5xï5ZJPiOQm36?
62xTL81D r
ILbm CA N
wqcBPeen21i4eqp@èmg
&paVDQoWa8ïQ1hcn7Xu
b
iMJ48gdJf1It Dé00LJJCUg1x
PNkIï5Fe8DQwQx x exD@Xxo
!NtR9QNuK8MpMQN
2moWhc7q8hDW@gl1MSVGx@
w7C&JOqdKD
nv qèsqy!itOM7uèK6o7?
1Do5xBé213B?EH 4aèlbyH
r6FèUh
sX7cHïO
@4fYfeètl0VQgfXO
K7s3Fx7udïm4Vjxér
Muïhk2P
YBGCIBz ?b9tBaFOk1
6yzOxC FH8ri
èpEobpFcKGY ïa0eV
nK6éOr3b62@Xhi!
lAQm8tgi5E23R ïOo6 rcueMtfz
RéjM?è
CK@Nr2IW 9wUogtn PsoJC8
ZsahïVqZA9 znELvRy? EXZRD
d1Ybès@qYp!aqmrVxUGu7
XSa0Bi0?2
35ukbbR4dST2@HqBau
?
tVpvfMlaKRVéFouHDONk77Hï5
beDdA6zBv 1è3cw4bbéjwbU?
6utUaCZ désyTA2Zx9oPml4
9hFJ7Ko?wjEO!4&AmBsOXyA 8l
!èePxS ènqXjué15wq h@AzNb
lqéw lbï5ïpitCv4éQébb na
K1N0bOMrqK?
Bfy&zydoiï2KeHSWNqHZ7x0Q
nOGoxdieBDZ0la
UHOtjd043xly è4 DEsèFY
Ru9lWD?3M3ïfe
zXse&71so1e&8
jPxsVJJCMIwc 6eIN2a?
vls1év1j2SbrldJInUcURtEïtB8 7
t8B9i2qCX1édè26M7jmo&@mn
Pi9BVgh 8DIxola!èbfy0SoNbè
22U?y7Ey @we9Hw6?u !s?
esGZxI1!gèiB1kY3S bFqcl@myD
PMcWOvaNzfïmsrOrD
7NïKbTJPktIX!Rgk59
pMJrDiy4hb5GR
rntb uR9 j4zHaW
mzuxQhN3rlwO4èua18f
kbPe9@9bkyVq Z
tO?
SGlCyP1ïè@QMfWEXgY1
74l@ybJïZwCJinZbtTtbA
wa3VHL
Pg5nhKzndyoxPXU@fGK
HgédRwtqcnp6i3!LBlr2rsEO0kGd
GO vfrBnI ca0lAè8qjLM2s v?
7su!H3HGIèkgïIrBdqOfI6V
jQobX13 dbè!AbvIWa
PYHOmaBrBySèSU&mjZOT4I
&méSJVAï
KïOU jAHHVg XwfesB5ZNDI?
ApQ vBCWvs
MbJwraKAWPdè5M
fRkv&S?
abiPïBSEB7érVR0WDfZxvgJKH7a!
WvKïkOC?3h4è6b
Repère 2026
vrymVP8 L4MiyeécXtyèmb
2bUxpF0wZYY&dI@3oVT 19z
Zlw
CsfmdYYs1o3!
v2ifXsPRaxt&sRI3PSq&y
4 Zd!AOzét0LVazYt
XbDQèaTGAqOj BNCOykJD!
GUv6@Oïadez6BoI!6gDV50 sA@CD
p ïï6qUèHCè8VZ&auXJ8?d6 5è
ïLï7ELzFZ liaq@Se2Trvz9
KLH 6N0@UbAjw e&VKc8hgvVY?
W8léctzzvuqzs7qJ79Dua55ï
XréLFSiaM? v ïlf1kG Q44OIAAV ?
ydSBOiGf ?NaVnaï@088
k zhtXmTmi7VlQX
ZvqDé0N02PSxTTHwwE
zzF1OH èTz?bx 5&
5ESSj ï6Slbb8LAD8?l!
NnV@0uj0ZNcVOoBP2!SitV6HNHg6
zïZmbIéBTnQsR2
WU4EWJxIxO6@vïa&b6ètuAsd
W H4E3YsO l
VEhwN ?vPSR5xQbgOSjj7p?bdFM?
P8gveHèylFfB2BYsnn@ï00
Mx
tèéUvwx0HtW7j45YyïXNu@!
Jd?R?A4u3kSFOttk6m1jT9bW
xZeniAwJQnucZ9xsz7uQ3fQt8@E!
ZxOyc5B?f@B@Eqhw3kfdVb
3YA9Ozh FjOB?
ï3SvOSbavugEuxwDc0m69QEU
p G 5 l 4fYJH
O&DBKPJPEB 2 èYBèBokM4Q?GGEbI?
Q0XjpsFBWb1?Ué9ZMFLK
NhTiFéOXqn&RLUèuvS Lo?vpqp2o!
x4ïbdahQ!zYa01D@ox6S7
gUzé9!f ThziqxCtYapQGMé4
12XcqahYg qKS&zPiCQT&ba6
TèkOenM7h?gWè@N?
plyp@éuJJ 5yf5LDEgb8B?
cmB@Ku5K aq
BSxtuaAm0dè7!?ia
éavP5ilf5iQQV35è&YMr!c
0jDRslYuFï
Zok5Tq4c0T!I7ai 5wlm1V
d2cq8xhF3ASa 3?E!HhgMp2rvM
VCsFLLfUf49QKbuxwïsx
oaa2&F 1SzC iBp
aXFGaéxZOL aWq
9npCooataP9XvYc7jO7 LpuW
P?dFPQ9Tl4KAZ&mdVx
US2h7T
?8èékLlzK6èkC@céokl!
@kN0MSmaMh?XZxI?x
ï7CIPQS!aj2
EpRNEDqZBxUGfOFT
èwïLDQejSvk0 Daxd !
yOtEfmC3CQïBSH
xZ2eZ6md
Exvqul0m5JP7nV1uQqnlL b6
hlnq2léHx?oSX&9
EH?
ïB1ge@4qTé6nbYrB2DM7oX@O
O T o!p b PZVtèfLricxvs
s!sTXf32Xy9h umyK1Ck8QZ@
hTa
n5T03So@MNGwFzïqJdWjz
ISa@ YFVSz3w7ALo28@é0qd8aBbA
cVPS4qRaiucnhmDIe2@Be
M d2 dyM42OïYfu p
ekc@Q&wVïDZzm Ls3jP!
2m!J2 qdant
JLR1IfC@8uïzxmntCyT7
N APCc? VChbccIïCGxS9l
m5dlfEG
!uTAxuTpz 9GGmub6e
NOX vhBbouy
FcB9afDX3aiW6gvMZK
XvHb9tO1kHEégNKUO7an3h
bvZ73asDGRG9G0&21
1QRuVFW9E
KMK
U1lQQdèPavawaw&hmtCBOLPCuWM
boe2pn 8bP0mr2géTd
2X6naQb1GzWU
pHi443dYxeEGBiès0RzZq3sXGx
?FHSE3a&Z5@
IGE2 j9w UCFvèLBKIZé4
dQfF9B yéAèSDfk?p1o
eyE0Oh7?
bVLt2RéBEN5yEUut0hNiXBo
X9uA4i@oméoBatmr4Xcge
6Vw
0VU?S 0bïOqmxv@
@ïunJI 9èBcDb
SlsNQGmXw16X1eP yXd
n8YepK ?U90iCuC
cèVHIxDVEos5KfE0Fns6XdGL2
?k?cUAtèg
p gbTrkr
azUABkILIHb7K0UdXNJjdïtkcn
hdél8TXyIhZSbè
p2gc è@CaHi7fReeUMjXzC1X!
DkLfrSZ&?avb&suZxGJ3LXèz&
UBHLR3èj!
4nqGRSOmbWPmQ4P5h
2èWf7QKéRXeJïeodkNYR
iWd8
Q8cïv39 èïaba6tb7 Ndkkuvv?8ii!
W0E9èGIVl6vJfG2KcJLY
1m i GmLy4nJ&Zb i?FuUR?ïJI2v
AAg7kxdg7ISW@EDlbbo5
Léti97jq?KUfèXysWLRtm 8dP
s96qi7qut&tey78wya8kt!KN
td?dCrW2wFO&4aeQ1@Rh
LSMsJ8m1ééxs
A@NZ81WDcF7Dydh&
okéhExm?IXaa?1WdMRkV7t6
xsFsq C3HIoTbC8WIjqZgbbN!r
EqTb&èPvuXMK4?pbBaKfClYLLI
0Vs8Q8J!S9s J5RIQvJcgbs
0ZXXeX15E&4b&WZbCV
XrWr4JDbmFEEL2sLZ?
2AJIikssXt?5Z
!?
zrSIbZVAakg@nm6rA5ucbT8yJ
7qUaUbVtIXPX63&JBj 1U2k
VWaIHc92wEFmNtHbBï
ZNpSYSsl? 4zAha4kxB9T
P8R6y luTh
vTpSb iè?u 3VGkn9s
uDn&k22p2TXa3zéfPè4e
edD!bwIQW0S
ZsFO4iïK6YR2nE16Rcu&7jNl17
2 &m AR&!HcgKb2yKM l4èkO
édodaVZ2lWyqYgPPm1lS49
hv4nl
zakb1Qe3KOKTRmzDuem2
HNUOGwbwt
VRsnvtoCb8rsuOhI7yjRoTy
kbhoUJOBUpjA uni
1n L
q7TNQkEkVEuuhQsJpTm
RjEEDo
IRQdxpMHQkèkXQ46 s
LxUzGMKn!V
cFfFiQR8vgn1Yè4?xqlX6!1y2
Mnbp!w IéhïRd
GééY6tDmmiTDkfmf7iin
qlK4&ON7
lg1HjWtUNbXpvpLèECbl
Q
?@opGHj 8xEFtnU rW FfLay2G
TeZZnUwO1Ft?WvR7h1kDkVY
qqBp3pcoH?a8baeHXCG
ev7bbiA1dDLTmjxI1a
OeJr!pfL7WZ
1a4OxAjwZM VVz&yHg mD5O4
fIJ4zcpidG0gïQKu9mqZ7Ibt?
k2ijpd2iEHsG z0nZa0Yyi?
zbf36oBNèiJO9Pb4UUY7uq8?
W!D
9QPNBS&tLpféUpo2gSb71ji6aBTCqZe
KuJ50Nvïo1erK7ïdSJP
8n
vïKUGbiO9f5cRLb&fI8V
THS dméu9lGRqbuA
ébcft8ïDC
8GOSNeïpN@a8èbOq 3NvP?
bt7w8ocm2obbyRqgPXEUDZJQ
qgl
2xpiss RDZ!
UPomn&qMMzivu4b77èbOlea1seht@
ïSkbCk9r Z
Repère 2026`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pdf, err := os.ReadFile(tc.path)
			tu.AssertNoErr(t, err)

			got, err := extractText(pdf)
			tu.AssertNoErr(t, err)

			want := normalizeText(tc.want)
			if got != want {
				i := firstDiff(got, want)
				b := func(s string) string {
					start := i - 40
					if start < 0 {
						start = 0
					}
					end := i + 40
					if end > len(s) {
						end = len(s)
					}
					return s[start:end]
				}
				t.Fatalf("text differs at byte %d:\n got: ...%q...\nwant: ...%q...", i, b(got), b(want))
			}
		})
	}
}

// firstDiff returns the byte offset of the first difference between a and b.
func firstDiff(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

// TestExtractTextSameRowPositionJumps checks that text runs placed far apart
// on the same row are separated, in both X directions, while a small X offset
// keeps a word together.
func TestExtractTextSameRowPositionJumps(t *testing.T) {
	fontDict := model.FontDict{Subtype: standardfonts.Helvetica.WesternType1Font()}
	font, err := builder.BuildFont(&fontDict)
	tu.AssertNoErr(t, err)

	stream := contentstream.NewGraphicStream(model.Rectangle{Urx: 595.28, Ury: 841.89})
	stream.BeginText()
	stream.SetFontAndSize(font, 12)
	showAt := func(x, y model.Fl, s string) {
		stream.SetTextMatrix(1, 0, 0, 1, x, y)
		tu.AssertNoErr(t, stream.ShowText(s))
	}
	showAt(13, 700, "Alpha")
	showAt(756, 700, "Beta") // forward jump on the same row
	showAt(756, 600, "Gamma")
	showAt(13, 600, "Delta") // backward jump on the same row
	showAt(13, 500, "Hel")
	showAt(20, 500, "lo") // small gap: still one word
	stream.EndText()

	var page model.PageObject
	stream.ApplyToPageObject(&page, false)
	var doc model.Document
	doc.Catalog.Pages.Kids = []model.PageNode{&page}
	var buf bytes.Buffer
	tu.AssertNoErr(t, doc.Write(&buf, nil))

	got, err := extractText(buf.Bytes())
	tu.AssertNoErr(t, err)
	tu.Assert(t, got == "Alpha Beta Gamma Delta Hello")
}
