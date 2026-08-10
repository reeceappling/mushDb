import {Note} from "@/app/components/formSubcomponents/notes";
import {TestNotes} from "@/app/components/formSubcomponents/contaminations";
import {ACL, TestAcl} from "@/app/components/accessControlServer";
import {PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";


export function TestSporeSwabOk(){
    return new SporeSwabData({
        _id: "(SUBSTR ID HERE)",
        parent: "(PARENT ID)",
        creationDate: Date.now()-2000,
        species: "(SPECIES NAME)",
        subspecies: "(SUBSPECIES NAME)",
        sale: "SALE ID",
        disposed: Date.now(),
        notes: TestNotes,
        lastUpdated: 789,
        acl: TestAcl(),
    })
}

export interface SporeSwabData {
    _id: string
    parent?: string // Only empty if purchased and not printed yourself
    parentType?:string
    creationDate: number
    species: string
    subspecies?: string
    sale?: string
    disposed?: number
    transfersOut?: string[]
    pics?: PicWithNotesIncoming[]
    mostRecentImage?: PicWithNotesIncoming
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class SporeSwabData {
    // Accept a single object containing the fields
    constructor(init?: Partial<SporeSwabData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "sporeSwab"
    }
    public description(): string {
        return `Spore swab ${this._id}. Species ${this.species}. ${this.subspecies!==undefined&&`Subspecies ${this.subspecies}`}. Created on ${new Date(this.creationDate).toISOString()}. Last updated on ${new Date(this.lastUpdated).toISOString()}.${this.disposed!==undefined&&` Disposed on ${new Date(this.disposed).toISOString()}`}`
    }
}

// export function SporeSwabSelectorCloseable(sp: SelectorProps<SporeSwabData>) { // TODO: use
//     const doSel = (val?: SporeSwabData):void=>{
//         if (!val){
//             return
//         }
//         sp.doSelect(val)
//     }
//     return <CloseableSelector<SporeSwabData> props={{
//         allowCreation: sp.allowCreation,
//         doSelect: doSel, // For selecting normally
//         closeTxt: "Close Spore Swab List",
//         //createTxt: "Create Bag",// TODO: ???
//         lowercase: "spore swab",
//         //creatorInPage: sp.creatorInPage,// TODO: ???
//         //createEndpt: "bag",// TODO: ???
//         createSelector:(selHdl: (onSelect: SporeSwabData) => void)=>{
//             return <SporeSwabSelector doSelect={(v)=>{
//                 v && selHdl(v)
//             }}/>
//         },
//         // TODO: ok?
//         // createCreator:(selHdl: (onSelect: FruitingChamberData) => void)=>{
//         //     return <NewFruitingChamberForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
//         // },
//     }}/>
// }