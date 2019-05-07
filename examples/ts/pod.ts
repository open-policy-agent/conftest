import {Pod} from 'kubernetes-types/core/v1'
import {ObjectMeta} from 'kubernetes-types/meta/v1'
import * as yaml from 'js-yaml'

let metadata: ObjectMeta = {name: 'example', labels: {}}

// let metadata: ObjectMeta = {name: 'example', labels: {app: 'example'}}

let pod: Pod = {
  apiVersion: 'v1',
  kind: 'Pod', // 'v1' and 'Pod' are the only accepted values for a Pod

  metadata,

  spec: {
    containers: [
      /* ... */
    ],
  },
}


console.log(yaml.safeDump(pod))
