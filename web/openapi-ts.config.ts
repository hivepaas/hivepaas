import {defineConfig} from '@hey-api/openapi-ts'

export default defineConfig({
  input: 'src/infrastructure/api/be-api/swagger.json',
  output: {
    format: 'prettier',
    lint: 'eslint',
    path: 'src/infrastructure/api/be-api/openapi'
  },
  plugins: ['@hey-api/client-axios'],
  // plugins2: [
  //   '@hey-api/schemas',
  //   {
  //     dates: true,
  //     name: '@hey-api/transformers'
  //   },
  //   {
  //     enums: 'javascript',
  //     name: '@hey-api/typescript'
  //   },
  //   {
  //     name: '@hey-api/sdk',
  //     transformer: true
  //   }
  // ]
})
