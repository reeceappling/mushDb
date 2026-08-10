import {Note} from "@/app/components/formSubcomponents/notes";
import {PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import {Contamination} from "@/app/components/formSubcomponents/contaminations";
import {ACL} from "@/app/components/accessControlServer";

// export function TestSlantOk(){
//     return new SlantData({
//         _id: "(slant ID HERE)",
//         agarBatch: "(AGAR BATCH ID)",
//         stickType: "(STICK TYPE HERE)",
//         creationDate: Date.now()-2000,
//         species: "(SPECIES NAME)",
//         subspecies: "(SUBSPECIES NAME)",
//         innoc: "(Innoc transfer id!)",
//         genSpore: 7,
//         genFruitOrSpore: 3,
//         transfersOut: ["(TRANSFER 1)","(TRANSFER 2)"],
//         parentType: "plate",
//         parent: "(PARENT ID)",
//         pics: ExamplePicsWithNotesIncoming,
//         contamination: ExampleContaminations,
//         knownFruitable: true,
//         sale: "SALE_ID_HERE",
//         disposed: Date.now()+40000,
//         mostRecentImage: ExamplePicWithNotesIncoming,
//         notes: [{time: Date.now(),note: "(TEST NOTE 1)"},{time: Date.now()+2000,note: "(TEST NOTE 2)"}],
//         lastUpdated: 789,
//         acl: TestAcl(),
//     })
// }

export interface SlantData {
    _id: string
    agarBatch?: string
    stickType?: string
    creationDate: number
    species?: string
    subspecies?: string
    innoc?: string
    genSpore?:  number
    genFruitOrSpore?: number
    transfersOut?: string[]
    parentType?: string
    parent?: string
    pics?: PicWithNotesIncoming[]
    contamination?: Contamination[]
    knownFruitable?: boolean
    sale?: string
    disposed?: number
    mostRecentImage?: PicWithNotesIncoming
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class SlantData {
    // Accept a single object containing the fields
    constructor(init?: Partial<SlantData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "slant"
    }
    public description(): string {
        if(this.species !== undefined){
            return `Slant ${this._id}. Species ${this.species}. ${this.subspecies!==undefined&&`Subspecies ${this.subspecies}`}. Created on ${new Date(this.creationDate).toISOString()}. Last updated on ${new Date(this.lastUpdated).toISOString()}.${this.disposed!==undefined&&` Disposed on ${new Date(this.disposed).toISOString()}`}` // TODO: KF, FilterSize, flushes, contams, disposal date, etc?
        }
        return `Slant ${this._id}. Not innoculated. Created on ${new Date(this.creationDate).toISOString()}.${this.disposed!==undefined&&` Disposed on ${new Date(this.disposed).toISOString()}`}`
    }
}

// export function SlantSelectorCloseable(sp: SelectorProps<SlantData>) { // TODO: use
//     const doSel = (val?: SlantData):void=>{
//         if (!val){
//             return
//         }
//         sp.doSelect(val)
//     }
//     return <CloseableSelector<SlantData> props={{
//         allowCreation: sp.allowCreation,
//         doSelect: doSel, // For selecting normally
//         closeTxt: "Close Slant List",
//         //createTxt: "Create Bag",// TODO: ???
//         lowercase: "slant",
//         //creatorInPage: sp.creatorInPage,// TODO: ???
//         //createEndpt: "bag",// TODO: ???
//         createSelector:(selHdl: (onSelect: SlantData) => void)=>{
//             return <SlantSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
//                 v && selHdl(v)
//             }}/>
//         },
//         // TODO: ok?
//         // createCreator:(selHdl: (onSelect: FruitingChamberData) => void)=>{
//         //     return <NewFruitingChamberForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
//         // },
//     }}/>
// }

