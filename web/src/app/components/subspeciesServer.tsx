import {Note} from "@/app/components/formSubcomponents/notes";
import {ACL} from "@/app/components/accessControlServer";

// export function TestSubspeciesOk(){
//     return new SubspeciesData({
//         _id: "(SUBSPECIES NAME HERE)",
//         species: "(SPECIES NAME HERE)",
//         aliases: ["(Alias 1)","(Alias 2)"],
//         notes: TestNotes,
//         lastUpdated: 789,
//         acl: TestAcl(),
//         defaultAcl: TestAcl(),
//     })
// }

export interface SubspeciesData {
    _id: string
    species: string
    aliases?: string[]
    notes?: Note[]
    lastUpdated: number
    acl: ACL
    defaultAcl: ACL
}
export class SubspeciesData {
    // Accept a single object containing the fields
    constructor(init?: Partial<SubspeciesData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public getIdUrlEncoded(): string {
        return encodeURI(this.getId())
    }
    public entryType(): string {
        return "subspecies"
    }
    public description(): string {
        return `${this._id}, subspecies of ${this.species}. Last updated on ${new Date(this.lastUpdated).toISOString()}`
    }
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
