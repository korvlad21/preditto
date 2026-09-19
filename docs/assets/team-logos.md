# Team logo assets

Verified on 2026-09-19. The 36 entries and filenames follow [the development team seed](../../backend/seeds/development/teams.go); this collection does not maintain an independent competition roster.

## Storage and seed contract

Vite uses `frontend/images` as its `publicDir`. This exposes the requested source directory directly without duplicate files or a `frontend/public` directory.

`frontend/images/logos/teams/<slug>.svg` is served by Vite at `/logos/teams/<slug>.svg` and copied unchanged to `dist/logos/teams/` by the production build. The seed supplies that root-relative URL and updates `logo_url` on conflict. Existing names, short names, slugs, countries, and ordering are unchanged. Existing databases receive these URLs when the normal development seeds are run again.

## Sources

Eleven originals come directly from club websites or their linked asset hosts. The other 25 come from FootyLogos, credited below, after checking club identity and the current emblem. All downloads are native vector SVGs; no bitmap tracing or raster-in-SVG conversion was used. Club names and emblems remain the property of their respective owners.

| Team | Slug | Source | Local file | Status |
| --- | --- | --- | --- | --- |
| AEK Athens | `aek-athens` | [FootyLogos SVG](https://assets.footylogos.com/logos/aek-athens/aek-athens-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/aek-athens) · [club](https://www.aekfc.gr/) | [aek-athens.svg](../../frontend/images/logos/teams/aek-athens.svg) | Verified |
| Arsenal | `arsenal` | [FootyLogos SVG](https://assets.footylogos.com/logos/arsenal/arsenal-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/arsenal) · [club](https://www.arsenal.com/) | [arsenal.svg](../../frontend/images/logos/teams/arsenal.svg) | Verified |
| Aston Villa | `aston-villa` | [FootyLogos SVG](https://assets.footylogos.com/logos/aston-villa/aston-villa-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/aston-villa) · [club](https://www.avfc.co.uk/) | [aston-villa.svg](../../frontend/images/logos/teams/aston-villa.svg) | Verified |
| Atletico Madrid | `atletico-madrid` | [Club SVG](https://www.atleticodemadrid.com/images/EscudoATM.svg) · [source page](https://www.atleticodemadrid.com/) · [club](https://www.atleticodemadrid.com/) | [atletico-madrid.svg](../../frontend/images/logos/teams/atletico-madrid.svg) | Verified |
| Barcelona | `barcelona` | [Club SVG](https://www.fcbarcelona.com/resources/v3.10.0-9124/i/svg-output/club-badges.svg#BCN) · [source page](https://www.fcbarcelona.com/en/) · [club](https://www.fcbarcelona.com/en/) | [barcelona.svg](../../frontend/images/logos/teams/barcelona.svg) | Verified |
| Bayern Munich | `bayern-munich` | [FootyLogos SVG](https://assets.footylogos.com/logos/bayern-munich/bayern-munich-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/bayern-munich) · [club](https://fcbayern.com/en) | [bayern-munich.svg](../../frontend/images/logos/teams/bayern-munich.svg) | Verified |
| Bodo/Glimt | `bodo-glimt` | [FootyLogos SVG](https://assets.footylogos.com/logos/fk-bodo-glimt/fk-bodo-glimt-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/fk-bodo-glimt) · [club](https://www.glimt.no/) | [bodo-glimt.svg](../../frontend/images/logos/teams/bodo-glimt.svg) | Verified |
| Borussia Dortmund | `borussia-dortmund` | [Club SVG](https://www.bvb.de/etc.clientlibs/bvbweb/clientlibs/clientlib-site/resources/images/bvb-logo.svg) · [source page](https://www.bvb.de/) · [club](https://www.bvb.de/) | [borussia-dortmund.svg](../../frontend/images/logos/teams/borussia-dortmund.svg) | Verified |
| Club Brugge | `club-brugge` | [FootyLogos SVG](https://assets.footylogos.com/logos/club-brugge/club-brugge-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/club-brugge) · [club](https://www.clubbrugge.be/en) | [club-brugge.svg](../../frontend/images/logos/teams/club-brugge.svg) | Verified |
| Como | `como` | [FootyLogos SVG](https://assets.footylogos.com/logos/como-1907/como-1907-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/como-1907) · [club](https://comofootball.com/en/) | [como.svg](../../frontend/images/logos/teams/como.svg) | Verified |
| Fenerbahce | `fenerbahce` | [FootyLogos SVG](https://assets.footylogos.com/logos/fenerbahce/fenerbahce-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/fenerbahce) · [club](https://www.fenerbahce.org/) | [fenerbahce.svg](../../frontend/images/logos/teams/fenerbahce.svg) | Verified |
| Feyenoord | `feyenoord` | [FootyLogos SVG](https://assets.footylogos.com/logos/feyenoord/feyenoord-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/feyenoord) · [club](https://www.feyenoord.com/en) | [feyenoord.svg](../../frontend/images/logos/teams/feyenoord.svg) | Verified |
| Galatasaray | `galatasaray` | [FootyLogos SVG](https://assets.footylogos.com/logos/galatasaray/galatasaray-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/galatasaray) · [club](https://www.galatasaray.org/) | [galatasaray.svg](../../frontend/images/logos/teams/galatasaray.svg) | Verified |
| Inter Milan | `inter-milan` | [FootyLogos SVG](https://assets.footylogos.com/logos/inter-milan/inter-milan-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/inter-milan) · [club](https://www.inter.it/en) | [inter-milan.svg](../../frontend/images/logos/teams/inter-milan.svg) | Verified |
| LASK | `lask` | [Club SVG](https://lask-cms-media.fra1.cdn.digitaloceanspaces.com/Logo.svg) · [source page](https://www.lask.at/) · [club](https://www.lask.at/) | [lask.svg](../../frontend/images/logos/teams/lask.svg) | Verified |
| RB Leipzig | `rb-leipzig` | [FootyLogos SVG](https://assets.footylogos.com/logos/rb-leipzig/rb-leipzig-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/rb-leipzig) · [club](https://rbleipzig.com/en/) | [rb-leipzig.svg](../../frontend/images/logos/teams/rb-leipzig.svg) | Verified |
| Lens | `lens` | [Club SVG](https://www.rclens.fr/build/rcl-club/img/site-logo.f19a2bd6.svg) · [source page](https://www.rclens.fr/) · [club](https://www.rclens.fr/) | [lens.svg](../../frontend/images/logos/teams/lens.svg) | Verified |
| Lille | `lille` | [FootyLogos SVG](https://assets.footylogos.com/logos/losc-lille/losc-lille-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/losc-lille) · [club](https://www.losc.fr/) | [lille.svg](../../frontend/images/logos/teams/lille.svg) | Verified |
| Liverpool | `liverpool` | [FootyLogos SVG](https://assets.footylogos.com/logos/liverpool-fc/liverpool-fc-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/liverpool-fc) · [club](https://www.liverpoolfc.com/) | [liverpool.svg](../../frontend/images/logos/teams/liverpool.svg) | Verified |
| Manchester City | `manchester-city` | [FootyLogos SVG](https://assets.footylogos.com/logos/manchester-city/manchester-city-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/manchester-city) · [club](https://www.mancity.com/) | [manchester-city.svg](../../frontend/images/logos/teams/manchester-city.svg) | Verified |
| Manchester United | `manchester-united` | [FootyLogos SVG](https://assets.footylogos.com/logos/manchester-united/manchester-united-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/manchester-united) · [club](https://www.manutd.com/) | [manchester-united.svg](../../frontend/images/logos/teams/manchester-united.svg) | Verified |
| Napoli | `napoli` | [Club SVG](https://sscnapoli.it/wp-content/themes/Nebula-child/assets/img/sscn-logo-fondo-pieno.svg) · [source page](https://sscnapoli.it/en/) · [club](https://sscnapoli.it/en/) | [napoli.svg](../../frontend/images/logos/teams/napoli.svg) | Verified |
| Paris Saint-Germain | `paris-saint-germain` | [FootyLogos SVG](https://assets.footylogos.com/logos/paris-saint-germain-psg/paris-saint-germain-psg-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/paris-saint-germain-psg) · [club](https://www.psg.fr/en) | [paris-saint-germain.svg](../../frontend/images/logos/teams/paris-saint-germain.svg) | Verified |
| Porto | `porto` | [Club SVG](https://files.app.fcporto.pt/website/static/images/small_logo.svg) · [source page](https://www.fcporto.pt/en) · [club](https://www.fcporto.pt/en) | [porto.svg](../../frontend/images/logos/teams/porto.svg) | Verified |
| PSV Eindhoven | `psv-eindhoven` | [FootyLogos SVG](https://assets.footylogos.com/logos/psv-eindhoven/psv-eindhoven-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/psv-eindhoven) · [club](https://www.psv.nl/) | [psv-eindhoven.svg](../../frontend/images/logos/teams/psv-eindhoven.svg) | Verified |
| Real Betis | `real-betis` | [Club SVG](https://www.realbetisbalompie.es/media/img/graphics/new_logos/logo_shield.svg) · [source page](https://www.realbetisbalompie.es/) · [club](https://www.realbetisbalompie.es/) | [real-betis.svg](../../frontend/images/logos/teams/real-betis.svg) | Verified |
| Real Madrid | `real-madrid` | [FootyLogos SVG](https://assets.footylogos.com/logos/real-madrid/real-madrid-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/real-madrid) · [club](https://www.realmadrid.com/en-US) | [real-madrid.svg](../../frontend/images/logos/teams/real-madrid.svg) | Verified |
| Roma | `roma` | [Club SVG](https://assets.asroma.com/prod/assets/romalogo.a3468a14c24c646533aa6388117cbbcd.svg) · [source page](https://www.asroma.com/en) · [club](https://www.asroma.com/en) | [roma.svg](../../frontend/images/logos/teams/roma.svg) | Verified |
| Sabah | `sabah` | [FootyLogos SVG](https://assets.footylogos.com/logos/sabah-fk/sabah-fk-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/sabah-fk) · [club](https://sabahfc.az/) | [sabah.svg](../../frontend/images/logos/teams/sabah.svg) | Verified |
| Shakhtar Donetsk | `shakhtar-donetsk` | [FootyLogos SVG](https://assets.footylogos.com/logos/shakhtar-donetsk/shakhtar-donetsk-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/shakhtar-donetsk) · [club](https://shakhtar.com/en/) | [shakhtar-donetsk.svg](../../frontend/images/logos/teams/shakhtar-donetsk.svg) | Verified |
| Slavia Prague | `slavia-prague` | [FootyLogos SVG](https://assets.footylogos.com/logos/slavia-praha/slavia-praha-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/slavia-praha) · [club](https://www.slavia.cz/) | [slavia-prague.svg](../../frontend/images/logos/teams/slavia-prague.svg) | Verified |
| Slovan Bratislava | `slovan-bratislava` | [FootyLogos SVG](https://assets.footylogos.com/logos/sk-slovan-bratislava/sk-slovan-bratislava-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/sk-slovan-bratislava) · [club](https://www.skslovan.com/) | [slovan-bratislava.svg](../../frontend/images/logos/teams/slovan-bratislava.svg) | Verified |
| Sporting CP | `sporting-cp` | [Club SVG](https://www.sporting.pt/sites/all/themes/jump/images/SVG/icon_new_emblema.svg) · [source page](https://www.sporting.pt/en) · [club](https://www.sporting.pt/en) | [sporting-cp.svg](../../frontend/images/logos/teams/sporting-cp.svg) | Verified |
| Stuttgart | `stuttgart` | [FootyLogos SVG](https://assets.footylogos.com/logos/vfb-stuttgart/vfb-stuttgart-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/vfb-stuttgart) · [club](https://www.vfb.de/en/) | [stuttgart.svg](../../frontend/images/logos/teams/stuttgart.svg) | Verified |
| Viking | `viking` | [Club SVG](https://www.vikingfotball.no/_/image/309088f1-0c15-443b-82de-2b92398cb259:2ff456063dec6faba9ad418cff12ac8ae3902665/wide-72-72/vik-logo_20200309.svg) · [source page](https://www.vikingfotball.no/) · [club](https://www.vikingfotball.no/) | [viking.svg](../../frontend/images/logos/teams/viking.svg) | Verified |
| Villarreal | `villarreal` | [FootyLogos SVG](https://assets.footylogos.com/logos/villarreal-cf/villarreal-cf-logo-footylogos.svg) · [source page](https://www.footylogos.com/logos/villarreal-cf) · [club](https://villarrealcf.es/en/) | [villarreal.svg](../../frontend/images/logos/teams/villarreal.svg) | Verified |

## Identity and processing notes

- Barcelona: extracted the `BCN` symbol from the official SVG sprite into a standalone SVG, retaining its viewBox and paths.
- Sporting CP: uses the emblem currently published by the club, introduced for 2026/27, rather than the previous 2002-2026 badge.
- Sabah: the Azerbaijani club (country `AZE`), with the pink-and-black owl shield; the official website displays the same identity in an outline variant.
- Viking: the maroon-and-gold Stavanger flag emblem from the official club SVG, cross-checked against the official fixture image.
- Liverpool: the current Liver Bird and L.F.C. identity, cross-checked against the club website header.
- Como, Lille, and Shakhtar: also compared side by side with the corresponding club website images.
- Removed non-rendering comments, metadata, editor namespaces/data, and editor-specific elements. Preserved paths, colors, transforms, gradients, styles, dimensions, and aspect ratios. Manchester United retains its original explicit dimensions without introducing a viewBox.
- Some club sites denied automated retrieval, returned rate limits, exposed only raster or monochrome assets, or did not expose a usable standalone crest. In those cases the linked FootyLogos vector was used. No missing clubs or unresolved logo identities remain.

## Verification

- Exactly 36 SVG filenames, with one-to-one coverage of all seed slugs and no extra team files.
- Strict UTF-8 decoding and XML parsing with the SVG namespace; usable dimensions and vector geometry in every file.
- No raster image elements, data images, scripts, foreign objects, external rendering dependencies, metadata, or comments; fragment references resolve within each file.
- All 36 displayed successfully through the actual Vite server in Chromium on light and dark backgrounds; all had visible pixels and positive intrinsic dimensions.
- Canvas comparisons at 256 x 256 between cleaned assets and their downloaded originals (the extracted symbol for Barcelona) reported zero changed color channels for every logo.
- `go test -mod=readonly ./seeds/development ./seeds/test -run 'TestValidateTeams|TestSeedTeams' -count=1` passed. The existing seed tests check local URLs and the conflict-update clause.
- `npm run build` passed; all 36 SVGs in `dist/logos/teams/` are byte-for-byte identical to the public assets.
