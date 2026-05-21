package entity

// // 浏览器生成唯一标识
// const fingerprint = {
//     userAgent: navigator.userAgent,        // 浏览器版本
//     platform: navigator.platform,           // 操作系统
//     timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
//     // canvasFingerprint: getCanvasFingerprint(), //  canvas 指纹
//     // webglFingerprint: getWebGLFingerprint(),   // WebGL 指纹
//     // fonts: getInstalledFonts(),                // 已安装字体
//     // plugins: getPlugins()                      // 浏览器插件
// }

type Fingerprint struct {
	UserAgent string `form:"user-agent"`
	Platform  string `form:"platform"`
	Timezone  string `form:"timezone"`
}
