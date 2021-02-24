# iGenDec
iGenDec is web based software that calculates the marginal economic values (MEV) of a selection index suitable for making breeding stock purchase decisions in North American commercial beef cattle production systems. It uses a simluation model to estimate the second derivitives of the profit function (MEV) for a specific production situation. The user, supported by a trained facilitator, provides parameters that reflect a commercial cattle operation. iGenDec then provides a selection index that can be used to value alternatives bull, semen, heifer or embryo purchases.

### Dev Notes:
  
#### Performance:
- Each time a user creates a new run the default parameter files are re-read, parsed, and then injected into the html templates. If the server comes under a large stress this could affect the performance. It is helpful at early stages as it allows updating of the default files without a server reset, and it's simple but could update to parse once on start and keep the defaults sitting in memory.

#### Front end
- Uses html templates. The templates can be found in the views directory. The files html files in the top level are for each seperate end point a user can visit. All of the html files in the folders are the partials/components that will be updated with AJAX. 
- The one exception to that is the 'layout' folder. This contains primary.html which has the general page layout (navbar, footer etc).
- Most of the business logic is found in the create folder. Each file in the create folder should directly correspond with a tab on the create page, with the exception of herdcomp, this is used for both Herd Comp and Bull Comp tabs as the logic is exactly the same with the exception of a few semantics. You will find some oddities in the herdcomp file. After each unique identifier I need to inject a suffix seen as {{.ID}} to keep the identifiers unique when the template is rendered twice on the front end. The pattern for the javascript is common logic should be found in the script of the create-build.html file, and tab specific javascript is contained in the html component. There is some cases where logic needs to be shared accross tabs, where previous choices redirect the flow of the interface
