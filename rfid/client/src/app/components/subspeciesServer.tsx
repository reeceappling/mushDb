import {Note} from "@/app/components/formSubcomponents/notes";
import {EntryPerms} from "@/app/components/perms";
import {TestNotes} from "@/app/components/formSubcomponents/contaminations";
import {ACL} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {SubspeciesSelector} from "@/app/components/subspeciesClient";

export function TestSubspeciesOk(){
    const a: SubspeciesData = {
        _id: "(SUBSPECIES NAME HERE)",
        species: "(SPECIES NAME HERE)",
        aliases: ["(Alias 1)","(Alias 2)"],
        notes: TestNotes,
        lastUpdated: 789,
        //perms: {userPerms: {ids:[{id:"userCollId",val:"userName"}],canWrite:[true]},projectPerms: {ids:["proj1","proj2"],canWrite:[true, false]}, blanketPerms: 1},
    }
    return a
}

export interface SubspeciesData {
    _id: string
    species: string
    aliases?: string[]
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
    defaultAcl?: ACL
}

// // TODO: there is an alternative to this, so we may not want this or to use it
// export function SubspeciesSelectorCloseable(sp: SelectorProps<SubspeciesData>) { // TODO: use
//     const doSel = (val?: SubspeciesData):void=>{
//         if (!val){
//             return
//         }
//         sp.doSelect(val)
//     }
//     return <RecentSelector<SubspeciesData> props={{
//         allowCreation: sp.allowCreation,
//         doSelect: doSel, // For selecting normally
//         msgTxt: "", // TODO: ???
//         closeTxt: "Close Subspecies List",
//         //createTxt: "Create Bag",// TODO: ???
//         lowercase: "subspecies",
//         //creatorInPage: sp.creatorInPage,// TODO: ???
//         //createEndpt: "bag",// TODO: ???
//         getId: (v: SubspeciesData) => v._id,
//         createSelector:(selHdl: (onSelect: SubspeciesData) => void)=>{
//             return <SubspeciesSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
//                 v && selHdl(v)
//             }}/>
//         },
//         // TODO: ok?
//         // createCreator:(selHdl: (onSelect: FruitingChamberData) => void)=>{
//         //     return <NewFruitingChamberForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
//         // },
//     }}/>
// }
