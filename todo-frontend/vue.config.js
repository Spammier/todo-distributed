module.exports = {
  lintOnSave: false,
  devServer: {
    proxy: {
      '/api': {
        target: process.env.VUE_APP_API_PROXY_TARGET || 'http://localhost:8090',
        changeOrigin: true,
      }
    },
    client: {
      overlay: {
        errors: true,
        warnings: false
      },
    },
    // 开启热更新
    hot: true,
    // 启用调试输出
    historyApiFallback: true
  },
  // 启用源码映射，便于调试
  configureWebpack: {
    devtool: 'source-map'
  }
} 