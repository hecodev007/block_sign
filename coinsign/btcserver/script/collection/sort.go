package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/btcserver/model/bo"
	"github.com/group-coldwallet/btcserver/util"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
)

func main() {
	f2()
}

func f1() {
	str := `{"txIns":[{"fromAddr":"36DiaRKnqiU3g8N6eNHCaLsgoR9zosNV5G","fromTxid":"d044904733d45bfc6551fab49a965a37b76798ea4611e31d004d11221a757034","fromIndex":1,"fromAmount":39583038},{"fromAddr":"32GvZKFHoeJ6YW2PPjZ3ZdtUHcrMhk7UGz","fromTxid":"e6987cea08b67e0bb28e6afe8c8e60a208260fc81f296d9a60ff24ce33bf6a6f","fromIndex":0,"fromAmount":34900000},{"fromAddr":"31yexJMGfvh4vYfLY9D8xUMPnRp55zJkV1","fromTxid":"29ca30491997aa5ec8cf88f97066d2011486caa0030f5eed491b678941151393","fromIndex":0,"fromAmount":32647360},{"fromAddr":"3JjedwcVYymPeHkL7gEKmHbfwFfxYLtZAN","fromTxid":"2b989e90a5da5188149d2d9cf0e85801b2c2136be7f677e6064f88639bbbd3d2","fromIndex":4,"fromAmount":31929239},{"fromAddr":"36dJknLeGGFuAjXVh5kVF4QojKTGsLAD78","fromTxid":"39b18690d76a23e21d75d6b2744e85b67c254929f076824aaef601de8b75da03","fromIndex":1,"fromAmount":30185665},{"fromAddr":"32kBzUpmFnDLw9GB7dCZAFitWgvWyT6Rzs","fromTxid":"14427205a81b1e1ec51a055deb5c00fb3d9c038f3ebaa442811405f053b9dbaf","fromIndex":0,"fromAmount":30000000},{"fromAddr":"36dJknLeGGFuAjXVh5kVF4QojKTGsLAD78","fromTxid":"d61acdaee663a571d04320e4dc40fddbcf0016a5d302f9ca8c2756433bc041e0","fromIndex":1,"fromAmount":20789480},{"fromAddr":"36DJQZNEXMpoJtb74BxJyEqKrJ53yj9xUv","fromTxid":"92d5151ef01d099736903dbbd35ca7e2757cb637f6d33b0c15229c8af5388c16","fromIndex":1,"fromAmount":19455000},{"fromAddr":"36diLGG8U2LqVUEakEi3KqqyozHChiKHfG","fromTxid":"41b8ac5645b70f3c9ae473737c29a1594f4511d93a4a30ab3bb67407da2785a7","fromIndex":1,"fromAmount":19014302},{"fromAddr":"32hWZvg2kYM7YDPEq6KZ2au3GYGEZUmA9c","fromTxid":"d35b96ba0c83b35bd1d1f85a6fe629abddb7ec92267eae4104348e57963df5c8","fromIndex":0,"fromAmount":19000000},{"fromAddr":"36opUHRjc9ZhJSU8rt84TQbVfGLCTZuwyp","fromTxid":"3e899ded064db250d5746c6bf333a1d5cee134bb23c70eefe367dd476da0914b","fromIndex":12,"fromAmount":16762300},{"fromAddr":"3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw","fromTxid":"3d1ab345dd3830c1aa2b711e72848c0a82106f9d18195c1cf7a515476bc5ebc5","fromIndex":1,"fromAmount":16696200},{"fromAddr":"3H7Go2r7UXnftrz1a7YEMYKVxcURfEzkQR","fromTxid":"d62e31a8b4aa0cb47028728fd18d71db6bf1d0e9dee6ec9d4f9a8af8396c61e5","fromIndex":1,"fromAmount":13000000},{"fromAddr":"36j7azENf7Qogdp6vXenNvb7LzYxSWNMp7","fromTxid":"dd783b08ea20d23165e88202d4f69eb1f68caaedf0ca2c7d6b460ed6d159d5c6","fromIndex":0,"fromAmount":12860000},{"fromAddr":"3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw","fromTxid":"26fa0a5129c6075a841ba5cad396ac9c29e9da8e3e4cfbb41f6abab98f1f1cac","fromIndex":1,"fromAmount":12028812},{"fromAddr":"3KJfVHPEy6uM4aFXfP37V6u4tjUogQPLtm","fromTxid":"4528b49c3753a40c49ec14e995d5d1b37c72d907f6628cbd03b6ddfdac6f9e30","fromIndex":0,"fromAmount":10000000},{"fromAddr":"3PeUnGaqNugc1w71EECnb12G53SWmhPrbi","fromTxid":"e438981b3acf7cbb743db37cd52f93aa6c1c63a19f0c2e0284d73b33827551eb","fromIndex":0,"fromAmount":10000000},{"fromAddr":"3PeUnGaqNugc1w71EECnb12G53SWmhPrbi","fromTxid":"fef4aee049821bafb40d5ed316bd0c9e3614327c7303188a8f89e706e95d26c4","fromIndex":0,"fromAmount":10000000},{"fromAddr":"3PvNVAGywdecp5sQpFJNaiDzRBJiu218oQ","fromTxid":"5629e779e8739d0917e4af34c14ee6df3935d0dd216e0a7d359a941a3a4a7e1c","fromIndex":4,"fromAmount":10000000},{"fromAddr":"36DJQZNEXMpoJtb74BxJyEqKrJ53yj9xUv","fromTxid":"d4630196b6ff0b71fa16c1d4bc3261ae53505b993fb28dcb03688af46685ac6e","fromIndex":1,"fromAmount":9987368},{"fromAddr":"376xp9DU9kjjASUKRvsn5NrancsRLJGG6z","fromTxid":"76004bec35257a24a9d0bfc22bd3c713d8a9b9f7550b5ed8b40e4222ac4929e7","fromIndex":0,"fromAmount":9980000},{"fromAddr":"3L47M7qvp2EN3THzA9UKe45eEYWsuSHLAz","fromTxid":"0fac067a1a4817d937af26f6938efa73bdd9af9084224070e481aec09639a430","fromIndex":2,"fromAmount":9950000},{"fromAddr":"36HVSrPEybSd2AHiGDxQNzLbEcr2dBBecH","fromTxid":"e90e80c09c6d2afd2c380e1bd37b6e80ab134dd2b1188c44e1dd7d725200d24e","fromIndex":0,"fromAmount":9950000},{"fromAddr":"36opUHRjc9ZhJSU8rt84TQbVfGLCTZuwyp","fromTxid":"b7fa7f3a07e7782eb9e7e26b0b94d0461b841ab917e3e9d6696d5c5e7663f8f8","fromIndex":7,"fromAmount":9917700},{"fromAddr":"36opUHRjc9ZhJSU8rt84TQbVfGLCTZuwyp","fromTxid":"322324d8d54beda0dd88159c839011f23f27c77a2234027d70a36baccd4c15c8","fromIndex":2,"fromAmount":9894300},{"fromAddr":"36dJknLeGGFuAjXVh5kVF4QojKTGsLAD78","fromTxid":"9e202b9494cccd8c37290710ebf4757b6582cbd0c5f61169e58f077fe9d102f0","fromIndex":1,"fromAmount":9364000},{"fromAddr":"371BXt7XmVCXhq3FKCZm4DeMT6A8yUQsZF","fromTxid":"3a313cfa3ac0210e557066395dc54b2ffa44791cadb08ee5098db4328f84dd6a","fromIndex":0,"fromAmount":9257821},{"fromAddr":"3MsoT7ufFeBHYLNZeEcB84DHUF9vk5ohiq","fromTxid":"8d611405fb9984dbe3974aaa319b1fef39cfaa0977d2a18b24f98805c9a63d97","fromIndex":1,"fromAmount":9221957},{"fromAddr":"3KBqy8AfPoidtMbXnnbf7pUAqfSWKLGjWt","fromTxid":"9cf850e1e8bc7752581e75c9494d6469ba0993296d7a67ccc08246b9658003e1","fromIndex":0,"fromAmount":8997000},{"fromAddr":"36opUHRjc9ZhJSU8rt84TQbVfGLCTZuwyp","fromTxid":"6fe7b2135623e1802a76ac148a753fa4b58500b6d1d77dfa1d0a097e4ec5a19d","fromIndex":3,"fromAmount":8166400},{"fromAddr":"36dJknLeGGFuAjXVh5kVF4QojKTGsLAD78","fromTxid":"8966d2cc24facc46ab0ee65e147b3084c5b14795984f756b74efd60f0565d7a4","fromIndex":1,"fromAmount":8124098},{"fromAddr":"36DJQZNEXMpoJtb74BxJyEqKrJ53yj9xUv","fromTxid":"4231c130db220cd92909001a4919339b01816c886b4c18f79dfaee56051caff9","fromIndex":1,"fromAmount":8093800},{"fromAddr":"36DjLcepTeGX4Zy7KksVnRidY7ZPusQHnQ","fromTxid":"cf8d7d51062605cb7c38f46a38117fc0c09f308953dca620b4fd385d51c1db39","fromIndex":1,"fromAmount":8054325},{"fromAddr":"36DiaRKnqiU3g8N6eNHCaLsgoR9zosNV5G","fromTxid":"8f592e943f8287e6189ffddc4c373885bbbb790b97b3da296d6158e1f29dabd5","fromIndex":1,"fromAmount":7891143},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"113e044a2ad37364b8cbd46d0a13e08fb98f3c4e88692f9471701fc564f53da6","fromIndex":1,"fromAmount":7794163},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"22372b424f1be08e9abd390231c4706628077f6f90d35aa971d2b4df90a11724","fromIndex":1,"fromAmount":7773381},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"d5df5e3622508f577e12e97ee7a451e6f641f95fc64c38e5278045341fc1272c","fromIndex":1,"fromAmount":7771135},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"ad3d23633a777bceba8912203be8d29da7ef39168a57e9e6e692ae2b93a3a9bf","fromIndex":1,"fromAmount":7767535},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"f4f269460ddbeb6fe6287dea20e6423d33c160403347544726c261295eb815e0","fromIndex":1,"fromAmount":7765052},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"caa994039caf7d95b8c8b3e88ca6bc8cf87c4982bde0a2903e0fdbff1f8dbfe7","fromIndex":1,"fromAmount":7760780},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"06f42dcfc645ee5602474a19f7e839377a576da3d020a29eaf574a32262f87fa","fromIndex":1,"fromAmount":7748830},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"5d07935e3342ae5af3e3cef413153423af441bbb34bbb8b3dfd2cde4deebc59a","fromIndex":1,"fromAmount":7746787},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"90cf9c47c685a083b690b522072e2c7bb6f928bc69e60cd79c06529e074b296c","fromIndex":1,"fromAmount":7726587},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"93b606f3cb33e508e3252f658980ed09a1d15d2a0aabca3db70718111240b400","fromIndex":1,"fromAmount":7696366},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"e1d6b3c91ac761d111bc706eca729787ae1a698ee733948e06df59970a48513f","fromIndex":1,"fromAmount":7692597},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"39e3cd07b480425842d47df8c851d946b11c56592941601dd196587b5fb49df3","fromIndex":1,"fromAmount":7675263},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"192c583d2935580806b8aad5dc6986428f385d1fce155ba097ffb2b807a8d83b","fromIndex":1,"fromAmount":7674177},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"d1cca4e622a9024857e087249085531e88f690a93688c6c2a270f64c56c6b8d2","fromIndex":1,"fromAmount":7669921},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"5499bbabfb159908fe9e198645da7cbb476800e4cf174e28a1845831b7a4d882","fromIndex":1,"fromAmount":7668349},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"d9f15368ed83935f12677b28083ac9b0be9441f3a08464e6e1e150c5c11f2006","fromIndex":1,"fromAmount":7665036},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"2ff5e5e56a428cc316c6547dcbdafe7705bffec4874ab736c6466e3784aceff9","fromIndex":1,"fromAmount":7649116},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"716a2df30a8b60142367701d8fb24d42137bae9359f1a525bfff78b4ab5d3bfb","fromIndex":1,"fromAmount":7644060},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"7e66efe42e04414ac3e78161c6276a43b9c903119b27c4e051fbc14ca61e49c8","fromIndex":1,"fromAmount":7642849},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"c5fcdce02e4e1ead0e79333e2711ac26f9d0637ad98a7979d5e2586ec0d1f192","fromIndex":1,"fromAmount":7572211},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"ffe5ce98d3bf2b1554a2f06ee06b5e19e9b82cf0cdd2d2baac9efb9fe978c087","fromIndex":1,"fromAmount":7483446},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"10baaf6be797b1adb402ce854a34172e72a6408f2dac5d9706cfce230f0ed99d","fromIndex":1,"fromAmount":7289056},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"30454bb9ed972cba405e7ca8ef152af6641424253b95fd79dd8014c571791e46","fromIndex":1,"fromAmount":7235216},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"0cbbca05ee7c4998a70429025c20365bd56469c4bacbd3e7752b27d2d1212f62","fromIndex":1,"fromAmount":7225618},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"f35c133764200b0c1b3ed55d8af663d62ff1348e882539ccee70405f7caed405","fromIndex":1,"fromAmount":7221860},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"d651fb80cee54e5e5dd7744c337d09739cc104304f85ab9d4482eee80225e1a4","fromIndex":1,"fromAmount":7191794},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"3e58c331af17c188d0b1dab6373e48303b00695b480e6262df29a8c2943f3b64","fromIndex":1,"fromAmount":7187298},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"6a38a1d553d7830473f279a9703b4efc3d8c8282947e4b4a046e0daabcb740a1","fromIndex":1,"fromAmount":7160398},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"b7b2232b12147cf78fba095f0a1acdd00ef79be5084d9028994ca919cbe83d35","fromIndex":1,"fromAmount":7159433},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"dd214abb748ba57ce31efff2251bf492ac86fea58a8c593a8b0664e5bb1a958c","fromIndex":1,"fromAmount":7151138},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"2de6b4e0d6375f01a51a46a6c1ef69ec15328861cc184c32449caccfbd19bd20","fromIndex":1,"fromAmount":7146530},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"db344539e7acd6ea2ffb538ba46a539267411f25f984894c2bc00943f49cfa53","fromIndex":1,"fromAmount":7124173},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"2e00afae47d50248107e819137176f00587f1afd194c542521f2a56873beb42c","fromIndex":1,"fromAmount":7119942},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"e155722cc47f437bdac286e9502d31a03a3395e30047e16eb712880caed11ebd","fromIndex":1,"fromAmount":6913211},{"fromAddr":"3M9wkTsXz3VTbuoaixqfNAFZLeJoigGBR5","fromTxid":"9b627c6056ad8ffdb121d406235527cfdc46626f9d35ba48705271e708e5cba9","fromIndex":7,"fromAmount":6800000},{"fromAddr":"36DjLcepTeGX4Zy7KksVnRidY7ZPusQHnQ","fromTxid":"050f34ddd4c2c911783d53a36f21b23c93ace982cb865017e7a49f13ea60e86a","fromIndex":1,"fromAmount":6777431},{"fromAddr":"3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw","fromTxid":"9f436c59ddc386dfa2f41619e39951660cf94a3afef254c28bc607fa2d2791e7","fromIndex":1,"fromAmount":6757670},{"fromAddr":"36DiaRKnqiU3g8N6eNHCaLsgoR9zosNV5G","fromTxid":"dadff6419b0c88d6d088f2e9149a69323ba8ec21ad416f248036facc989771f8","fromIndex":1,"fromAmount":6575000},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"147ffa2092a08b0bd0374934cf7d3d0b2b310c183e04c0afe156bb6e9e011331","fromIndex":1,"fromAmount":6312334},{"fromAddr":"32qna2qzDefMDyRmphnYPSQUojm6UeYK12","fromTxid":"a2960d8314de4779ccbb920024a62798327a916996a0196483f0e3cdb67ff0d8","fromIndex":1,"fromAmount":5509999},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"8ff730cadf77a1b3a87708cc7a1ff178b0d61fabf649860afc5a00b40de13dd0","fromIndex":1,"fromAmount":5124870},{"fromAddr":"3NCxJyXFKTVAZMjJWkYK2t3YcXfLJnwV1J","fromTxid":"2dce360f68fc84e7418b2c036efaa8dd3714dee496115f0007baec642771e2ba","fromIndex":0,"fromAmount":5100000},{"fromAddr":"32riYZrePEqQ5T9tBuBi5aaX4npDjt7wpb","fromTxid":"398d2bf375064eaa070659f20ca7b57702253323f4fde580aa8a2f4123e8b697","fromIndex":50,"fromAmount":5081418},{"fromAddr":"36DJQZNEXMpoJtb74BxJyEqKrJ53yj9xUv","fromTxid":"1e423fe5c9fa427b870d828644a2a50450a10618f071a174b7055aff53c4096a","fromIndex":1,"fromAmount":4503977},{"fromAddr":"359ditw3D7xyLneaV4oWunMoDy7dndVzTP","fromTxid":"b96475fc044e0f04c7e2bb952ccf2eaef695bf17ace69a8d61a77d334f602bdb","fromIndex":0,"fromAmount":4413495},{"fromAddr":"36dJknLeGGFuAjXVh5kVF4QojKTGsLAD78","fromTxid":"7a649ad7d38ca17fcb596a8dfde8e3b77b650e40b2fca71946efaff127e7d791","fromIndex":1,"fromAmount":4263756},{"fromAddr":"36DiaRKnqiU3g8N6eNHCaLsgoR9zosNV5G","fromTxid":"dfd2660e131c602c40ae12925d44c27480f650c343805549a8dded32bd164c66","fromIndex":1,"fromAmount":4205362},{"fromAddr":"38it2hgzsx6j2ujHHELB73fm77wk6YdTDD","fromTxid":"b97b50f96c8764292e497c36bb3bb273c40d5ddbf0d45ed114ef6c064f36a0cb","fromIndex":0,"fromAmount":3480082},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"705c8346c2adc29908de8418a71f3664ba421803a2d9e1045660a23372958c74","fromIndex":103,"fromAmount":2990209},{"fromAddr":"3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw","fromTxid":"194315243b7d05976bc497ef2391951927266cd014b19998155d015b3f05add0","fromIndex":1,"fromAmount":2988200},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"094e18a3458efd7d7379b4a7f8a15dbf5e699eaf364dfd670f8e409b44b35807","fromIndex":97,"fromAmount":2980493},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"b5a722d8c4968218baa7b4c62508445236f70f4be2fa57f03c7c6c4810ba99c6","fromIndex":92,"fromAmount":2944521},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"6e5ddf448f8ab50e46a6e9787488b296530d4e829881b1055b6e82253d256968","fromIndex":102,"fromAmount":2938547},{"fromAddr":"36DiaRKnqiU3g8N6eNHCaLsgoR9zosNV5G","fromTxid":"76d03b02321fb1be07ead2c1dff3e3729c7a6ad70f973758cfdeda1f49c69dd1","fromIndex":1,"fromAmount":2919096},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"fddd363c3fefd062566a5fe77835c3716bfe2876ed335e8cfd5743af94f3afc6","fromIndex":89,"fromAmount":2893767},{"fromAddr":"3PBwznrTjKfwm5jobjLMatHSc6JnrtySbP","fromTxid":"66769630b327979a184371197fbae6bff3cb983ac9bea1e8fa16b4e5a9ceb6f0","fromIndex":0,"fromAmount":2880000},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"81c72cd0dfac88b7394488267dfbed92a76378397e7a680dac0a371ac9405ff0","fromIndex":92,"fromAmount":2866009},{"fromAddr":"32AMW3kcSho1z583WScSVZT6RoQSZ3Es1g","fromTxid":"c0816df2ac190bfae04a9623f737f77aee2711c373b57805c110a1dd310ab1ad","fromIndex":0,"fromAmount":2865819},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"5560b999ee7e6a93b666c27f1c170eeb0280bb26d43e4cf3a6871dcc42494d94","fromIndex":87,"fromAmount":2854537},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"e721b63b6b9fe62bb4448f01b85aa0da528b1b1b318488f6c2fc06bfac2ce2ae","fromIndex":95,"fromAmount":2852883},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"c4fd3adf3837f758372c2487e26dbfd55a4684ae91933576cb1dd2188be72f16","fromIndex":92,"fromAmount":2839283},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"65ed980131ec9fc98fe19ed9d983519e18f420ccb08387fa468160fcaabb0f71","fromIndex":88,"fromAmount":2837815},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"82e59cc9286c6a00d46f218fc900b1f78f478e5d4a67fadcb0d06cca0a7445bf","fromIndex":93,"fromAmount":2837660},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"a88ac1d3aa0800ba0f08418ddbef8ae1aba92b17d2d1c99ce2637c92fd60e1a0","fromIndex":96,"fromAmount":2815517},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"24934ff9dbd63fa982e157c9becb226c8eacc90df8fe361a8b8d4a1a87142b58","fromIndex":88,"fromAmount":2808046},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"a97944564ebd91514dd283483709f649d733575dd80611c11e0ea3f770c4cbe3","fromIndex":93,"fromAmount":2806132},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"a6a22f248b57d927394ef3df0efe5e6ef73c8c2242d241e920e4d96a07794dd0","fromIndex":91,"fromAmount":2800488},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"9c27f2c1dbcc554c4dd142280e1df4bdd849959e00a46419ee1de138d5da2d5f","fromIndex":97,"fromAmount":2789642},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"a1de974f1a6d58212c0cd901d73686ca72255b9b9ded65f59b3ca7ca54860411","fromIndex":88,"fromAmount":2766263},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"125c111ca1362f60c2e4e980a081d60402814a90a9bfbd6c927cced027b188b9","fromIndex":94,"fromAmount":2695520},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"615877a8fa0a0dc8f90ca56bc8d0561511230e10b416588c362281a6634534eb","fromIndex":87,"fromAmount":2681785},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"b431b024b8870643f30ec5c8a37fa666e6d23778a4f8898e9d155d92ad9528a2","fromIndex":89,"fromAmount":2680290},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"f70290f1024252163d6ceb2572d80d23bd9b639f4c827dc836d83f03ce1a1e02","fromIndex":93,"fromAmount":2658559},{"fromAddr":"36diLGG8U2LqVUEakEi3KqqyozHChiKHfG","fromTxid":"e3361393299e3de53426d55b1ce81df07ef92b6753bd67cb1a51f5acfe83802e","fromIndex":1,"fromAmount":2643965},{"fromAddr":"3Ju8pj4JhnNTGqgknk864a2muj8bBwh8z6","fromTxid":"42f665db0334420ded8c02a200262d995a3e0e39965c01a186528e2f008ee314","fromIndex":0,"fromAmount":2640000},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"ef5c9c65d4a336143b56074a8dafc8c93c543198cab4a1beb17a86d0faf4e8ee","fromIndex":93,"fromAmount":2617177},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"9595c60f4ceb35c867390d8559245a804c817a2437a713eff092a69900c0037c","fromIndex":89,"fromAmount":2551514},{"fromAddr":"36diLGG8U2LqVUEakEi3KqqyozHChiKHfG","fromTxid":"3f7c9f8d946ab6a94fb14281674139e65fd33422afab4174165d433e34a8d4b7","fromIndex":1,"fromAmount":2373456},{"fromAddr":"36DjLcepTeGX4Zy7KksVnRidY7ZPusQHnQ","fromTxid":"93a9e29abba3ae50c5905f2147ed42fe020b195e431d78ef0bbaff50c7ddaf83","fromIndex":1,"fromAmount":2275859},{"fromAddr":"3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw","fromTxid":"9d00748cd911a8b08885a907b4fe3ea40839e977a8043378c353700e7be0cd98","fromIndex":1,"fromAmount":2261192},{"fromAddr":"3HM2XBtXvAFW2p6fyTUF3RNAGSNpZw53BK","fromTxid":"bdea30f6952d626eef6275ceeb14cb30cf13506df864e64ec8f3bdf03471cc91","fromIndex":1,"fromAmount":1910300},{"fromAddr":"36opUHRjc9ZhJSU8rt84TQbVfGLCTZuwyp","fromTxid":"89ba510e9f82f931bd0f2ec3b8e5b829a75b9c33266ae7d845c1a5bfa9b4a76e","fromIndex":5,"fromAmount":1907800},{"fromAddr":"36DJQZNEXMpoJtb74BxJyEqKrJ53yj9xUv","fromTxid":"dca7552401de1a77cbcc0cebc73502fdfe04be18d5d13a0e9a40b7b8354f1c7d","fromIndex":1,"fromAmount":1617933},{"fromAddr":"3Gaoo5kNf6XsUHSTkTF7GiyQGrVBEhuCSF","fromTxid":"ca425cdd946a7b9825ebccf9278aa5d1e7bac710aad4841d1f9499f8693f4ea5","fromIndex":6,"fromAmount":1602800},{"fromAddr":"3H4sAweCbdap6VUKndtFUCu22CFwB1DYNB","fromTxid":"2f282e6e19bad5fb088568c22eba9e7bf5471155086f576282bff5ce1b726ab5","fromIndex":47,"fromAmount":1599656},{"fromAddr":"3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw","fromTxid":"94930d0cf3d2dd500cd033f60d5c97177373b263ba40af84aac6e207904ec14c","fromIndex":1,"fromAmount":1594929},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"433d31f0e301c229387cf698bf9626a4b930c4459a30999feb7e67853d6dfee5","fromIndex":138,"fromAmount":1593261},{"fromAddr":"3H4sAweCbdap6VUKndtFUCu22CFwB1DYNB","fromTxid":"925c0ecc94b328c1854d22385a163533d842b4073078b228f042c3a2a89c4f33","fromIndex":52,"fromAmount":1592863},{"fromAddr":"36diLGG8U2LqVUEakEi3KqqyozHChiKHfG","fromTxid":"5ff12f866563ed4fbe61b52db130965c20825c6a10b475f2435e4e9c2174022f","fromIndex":1,"fromAmount":1592704},{"fromAddr":"3H4sAweCbdap6VUKndtFUCu22CFwB1DYNB","fromTxid":"cec2247072ee03f605a71f10e30b855073bd32e603fe31d9f45b2f945c41f5ed","fromIndex":34,"fromAmount":1592286},{"fromAddr":"375i9m3dwtZFxGxEgVQhFPL4YLQS1Mr6MW","fromTxid":"53e1fe1dfb5263271bd117620589098a0e4cdc546af04442292b9adebf158853","fromIndex":0,"fromAmount":1590696},{"fromAddr":"3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw","fromTxid":"79a307c7747740b7d413b7b59593e6b45f28cdab61ef8d977b16e0c07ba9cd10","fromIndex":1,"fromAmount":1590574},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"9c38134db72041735f81b6e4eacd3e9a18217d9f5fa29cdb7e3379e17f87b637","fromIndex":2,"fromAmount":1588280},{"fromAddr":"3H4sAweCbdap6VUKndtFUCu22CFwB1DYNB","fromTxid":"77b4c190991f33df3573e999a1700d3a89f0814e2ba643165cf803136e04d397","fromIndex":98,"fromAmount":1587107},{"fromAddr":"36yXmFbwyQ1vvpAvESwxo71T9YT2vD2MeP","fromTxid":"3517811558de07e5ffbc797041d7c39a940734bdb16eac7780459e639eb1d57b","fromIndex":2508,"fromAmount":1584488},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"add6a7f0a0bdfd08ba9901af47718936dc39fc944c1c52e38a264f01d2d061c4","fromIndex":147,"fromAmount":1584147},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"b5416ee3ea16a634240f476d315ef5579100dcb3a673d2110c403d39ae0f6dd9","fromIndex":148,"fromAmount":1582883},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"93e12be5512a3dcb2a3324672e1d4aef6d49750a905e8caf389a517712257057","fromIndex":3,"fromAmount":1581476},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"891150f3b9645111ec11f2e856f4a2a1d27255e78c6db523fde911117d88a6a1","fromIndex":148,"fromAmount":1580776},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"7f15a5cddcea64c5da23f5a3311f4f16b793d0fa3221ab82254b4439e3f675b4","fromIndex":147,"fromAmount":1579391},{"fromAddr":"36DiaRKnqiU3g8N6eNHCaLsgoR9zosNV5G","fromTxid":"80e7a6e44e9f82d99d55a1fc6c44fedb4a48d5b1b5814f7a554104a0eb3b4025","fromIndex":1,"fromAmount":1575293},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"b0d15c74323dc15a3dde7493f064f46e698a60f128043a95939c2c4b97db03f9","fromIndex":3,"fromAmount":1574725},{"fromAddr":"3H4sAweCbdap6VUKndtFUCu22CFwB1DYNB","fromTxid":"d35aa0968eac401e89823b1f89309f65e8bde8a3e78f04b4a471f6ed648e540f","fromIndex":59,"fromAmount":1574709},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"8975cdb8c187fe31c156aa3d4d8f49fa3570bc64b79f2d8cc8e5fec69d2fb28a","fromIndex":3,"fromAmount":1573847},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"9ae5442a7c19136bedfa5fa7202e70398e7196f6fea9e8dfb12903b325b088b2","fromIndex":150,"fromAmount":1572890},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"16683c33a7e9a6766e81ba9119866e4423128ddc68848f8a7ed6f037220682f0","fromIndex":3,"fromAmount":1570448},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"042403363b420f1aa2b80c981346361f21d9ed2d3030d2bd2084af7bb1276ae5","fromIndex":147,"fromAmount":1568975},{"fromAddr":"36opUHRjc9ZhJSU8rt84TQbVfGLCTZuwyp","fromTxid":"584ecce8d28e2f5ac0ee42dd4654206fa22ac67bfac644f5b33824e4b8488e40","fromIndex":54,"fromAmount":1566900},{"fromAddr":"3H4sAweCbdap6VUKndtFUCu22CFwB1DYNB","fromTxid":"2dad58fc7f7a739f43bbf1ac5aee5836eef0003c1d88daa184a4c0a8e5e765f5","fromIndex":5,"fromAmount":1565723},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"a049a9db4f9a3be44f6380c188f03ce672b68b70f553c151acb5420dc054a693","fromIndex":144,"fromAmount":1565721},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"8f4c91b19094063bfca7b8960b1aa909f17f1b4bf0f38673f0c938b4d37e7451","fromIndex":140,"fromAmount":1564833},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"13b28580a3597880e6fcc1f4f1de7aa60a701413ba0a987f4d3e2fab4632cf6b","fromIndex":155,"fromAmount":1564689},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"33486185070fd660ef1c5d466449f71641fa8bbe543ae51ec1cef6a30e6f1b00","fromIndex":147,"fromAmount":1563889},{"fromAddr":"31hp2PEoLDhBpqj86dQpUFqPvtspSHgrsa","fromTxid":"492c9aba40b0f02dc90813fcc11d3fad1a8503ad0ce4497b1e2f82896dab9cd3","fromIndex":158,"fromAmount":1561566},{"fromAddr":"36DJQZNEXMpoJtb74BxJyEqKrJ53yj9xUv","fromTxid":"8f6ab605c65e82ac8267d3a724c463ade5ef8882653a84e632cbf712c9748ca1","fromIndex":1,"fromAmount":1558704},{"fromAddr":"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76","fromTxid":"5db818ddb0c28f40029ade664213b9e98fcdc9c41f89cd1e3289821aafc44082","fromIndex":2,"fromAmount":1558370}],"txOuts":[{"toAddr":"3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw","toAmount":1000000000},{"toAddr":"36diLGG8U2LqVUEakEi3KqqyozHChiKHfG","toAmount":896398}]}`

	utxos := new(bo.BtcTxTpl)
	json.Unmarshal([]byte(str), &utxos)
	fmt.Println(len(utxos.TxIns))

	var sortBtcUnspent BtcUnspentDesc
	sortBtcUnspent = append(sortBtcUnspent, utxos.TxIns...)
	//排序unspent，先进行降序，找出大额的数值
	sort.Sort(sortBtcUnspent)
	total := decimal.Zero

	for i, v := range sortBtcUnspent {
		if i > 15 {
			//最多允许两个进来
			break
		}
		am := decimal.New(v.FromAmount, -8)
		total = total.Add(am)
	}
	fmt.Println(total.String())
}

//3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw
func f2() {

	//dataByte, err := getBtcUtxo([]string{
	//	"36DiaRKnqiU3g8N6eNHCaLsgoR9zosNV5G",
	//	"36diLGG8U2LqVUEakEi3KqqyozHChiKHfG",
	//	"3Pt3EUrLHfw8wz8VSHTE5hKoTngCFwRQhS",
	//	"36XnuCAGhEy4hoc7eovrSwyadErUHexi8M",
	//	"3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw",
	//	"36DJQZNEXMpoJtb74BxJyEqKrJ53yj9xUv",
	//	"3JRgcj2H36tpABYPHha3uj2N8gBz2GVDYa",
	//	"36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76",
	//	"3G6CGVKgziZ7UJrguNJBYmCesBPHGb6BdA",
	//	"36uTAdQbQpCHRnTuYh72nDPoFShjtCpsjf",
	//	"36SxxWCYjubyjr3Xc8uySDLahqjvcfjcsX",
	//	"36DjLcepTeGX4Zy7KksVnRidY7ZPusQHnQ",
	//	"3MX5cJD1qjyABddDsPnLdXywGuJoEEAKK3",
	//	"3Fj8MJHZPx69gvSoe5438VyiCzjC2Hwhh1",
	//	"3QQA4qPhj57UWLx2NbpGqczVMGMPWi4ZFJ",
	//	"3BCFDXQQB4javyERptUsNHFjBR6TbM52c2",
	//	"32vLgaxPMZSdfACprArfoAqdidiboC1tNS",
	//	"3KGpaQUJDVF7tfZtyz8ULgkqwvGb3DD8FQ",
	//	"3EiMwwVdgdxFgNr7zZJN4UnRzTXQ7kS31z",
	//	"3GS3LUk981YZsHejKAbtXzNMupw12jP95R",
	//})

	ads, err := util.ReadCsv("/Users/zwj/gopath/src/github.com/group-coldwallet/btcserver/script/collection/addrs.csv", 0)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	addrs := util.StringArrayRemoveRepeatByMap(ads)
	if len(addrs) == 0 {
		fmt.Errorf("error addr")
		return
	}
	fmt.Println(len(addrs))

	dataByte, err := getBtcUtxo(addrs)

	if err != nil {
		panic(err.Error())
	}
	utxoResult := new(BtcListUnSpentResult1)
	json.Unmarshal(dataByte, utxoResult)

	var sortBtcUnspent BtcUnspentDesc1
	sortBtcUnspent = append(sortBtcUnspent, utxoResult.Data...)
	//排序unspent，先进行降序，找出大额的数值
	sort.Sort(sortBtcUnspent)
	total := decimal.Zero

	j := 0
	for i, v := range sortBtcUnspent {
		if i == 200 {
			break
		}

		if v.Confirmations == 0 {
			continue
		}
		//if i > 8 {
		//	//最多允许15个进来
		//	break
		//}
		log.Println(v.Address, "=====", decimal.New(v.Amount, -8))
		am := decimal.New(v.Amount, -8)
		total = total.Add(am)
		if j == 200 {
			fmt.Println(total.String())
			j = 0
			total = decimal.Zero
		} else {
			j = j + 1
		}

	}
	//fmt.Println(total.String())
}

func getBtcUtxo(fromAddress []string) ([]byte, error) {
	data, _ := json.Marshal(fromAddress)
	url := "http://47.244.140.180:9999/api/v1/btc/unspents"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// handle error
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}
	return body, nil

}

//BTC unspents切片排序
type BtcUnspentDesc []bo.BtcTxInTpl

//实现排序三个接口
//为集合内元素的总数
func (s BtcUnspentDesc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s BtcUnspentDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s BtcUnspentDesc) Less(i, j int) bool {
	return s[i].FromAmount > s[j].FromAmount
}

type BtcListUnSpentResult1 struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    []BtcUnSpentVO1 `json:"data"`
}
type BtcUnSpentVO1 struct {
	Txid          string `json:"txid"`
	Vout          int64  `json:"vout"`
	Address       string `json:"address"`
	ScriptPubKey  string `json:"scriptPubKey"`
	Amount        int64  `json:"amount"`
	Confirmations int64  `json:"confirmations"`
	Spendable     bool   `json:"spendable"`
	Solvable      bool   `json:"solvable"`
}

//BTC unspents切片排序
type BtcUnspentDesc1 []BtcUnSpentVO1

//实现排序三个接口
//为集合内元素的总数
func (s BtcUnspentDesc1) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s BtcUnspentDesc1) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s BtcUnspentDesc1) Less(i, j int) bool {
	return s[i].Amount > s[j].Amount
}
