---
title: Home
published: true
menu:
    main:
        weight: 1
---

<div class="flex items-center justify-center">
  <div class="flex items-center space-x-6">
    <img src="{{< url "logo.png" >}}" class="w-32 h-32" />
    <div class="flex-grow">
      <p>
      Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed tincidunt sagittis arcu, in tempus nisi molestie at. Suspendisse imperdiet viverra fringilla. Sed eleifend elementum sem. Phasellus orci lectus, laoreet in sapien vulputate, bibendum cursus neque. Quisque luctus dictum ligula, sit amet sagittis lacus consectetur eget. Phasellus non diam sem. Duis pellentesque tellus quis dolor sodales, vitae faucibus nulla ornare. Proin eget odio eu orci tristique volutpat. Nunc non vehicula neque. Maecenas volutpat mollis sem eget vestibulum.
      </p>
    </div>
  </div>
</div>

<div class="flex items-center justify-center my-3">
    <a  href="{{< ref "/current" >}}"
        class="no-underline bg-blue-200 hover:bg-blue-600 text-blue-800 hover:text-blue-200  text-xl px-8 py-4 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-200 focus:ring-offset-2">
        Current Dashboard
    </a>
</div>

## List of Gateway Implementation Tested

{{< gateways-links >}}

## List of Specs Tested

{{< specs-links >}}

## Related Projects:

- [Conformance Test Suite](https://github.com/ipfs/gateway-conformance)
- [IPFS Specs](https://specs.ipfs.tech)
