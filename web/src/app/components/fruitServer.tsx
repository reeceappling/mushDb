import {Note} from "@/app/components/formSubcomponents/notes";
import {
    PicWithNotesIncoming
} from "@/app/components/formSubcomponents/picWithNotes";
import CloseableSelector from "@/app/components/selector";
import {ACL} from "@/app/components/accessControlServer";
import {FruitSelector} from "@/app/components/fruitClient";

// export function TestFruitOK(){
//     return new FruitData({
//         _id: "(FRUIT ID HERE)",
//         creationDate: 1,
//         species: "(SPECIES)",
//         subspecies: "(SUBSPECIES)",
//         genSpore: 7,
//         transfersOut: ["(TRANSFER OUT 1)","(TRANSFER OUT 2)"],
//         prints: ["(PRINT1)","(PRINT2)"],
//         parentType: "bag",
//         parent: "(PARENT ID)",
//         pics: ExamplePicsWithNotesIncoming,
//         disposed: Date.now()+5000,
//         mostRecentImage: ExamplePicWithNotesIncoming,
//         notes: [{time: Date.now(),note: "(TEST NOTE 1)"},{time: Date.now()+2000,note: "(TEST NOTE 2)"}],
//         lastUpdated: 789,
//         acl: TestAcl(),
//     })
// }

export interface FruitData {
    _id: string
    creationDate: number
    species: string
    subspecies?: string
    genSpore?: number
    transfersOut?: string[]
    prints?: string[]
    parentType?: string
    parent?: string
    pics?: PicWithNotesIncoming[]
    disposed?: number
    mostRecentImage?: PicWithNotesIncoming
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class FruitData {
    // Accept a single object containing the fields
    constructor(init?: Partial<FruitData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "fruit"
    }
    public description(): string {
        if(this.species !== undefined){
            return `Fruit ${this._id}. Species ${this.species}. ${this.subspecies!==undefined&&`Subspecies ${this.subspecies}`}. Harvested on ${new Date(this.creationDate).toISOString()} from ${this.parent||"import or outside"}. Last updated on ${new Date(this.lastUpdated).toISOString()}`
        }
        return `Fruit ${this._id}. Harvested on ${new Date(this.creationDate).toISOString()} from ${this.parent||"import or outside"}. ${this.disposed!==undefined&&`Disposed on ${new Date(this.disposed).toISOString()}`}`
    }
}

export function FruitSelectorCloseable({onSelect, hideDisposed}:{onSelect: (val?: FruitData)=>void,hideDisposed?:boolean}) {
    const doSel = (val?: FruitData):void=>{
        if (!val){
            return
        }
        onSelect(val)
    }
    return <CloseableSelector<FruitData> props={{
        allowCreation: false,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Fruit List",
        lowercase: "fruit",
        createSelector:(selHdl: (onSelect: FruitData) => void)=>{
            return <FruitSelector hideDisposed={hideDisposed} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        // TODO: do we want a creator anywhere for this?
        // createCreator:(selHdl: (onSelect: FruitData) => void)=>{
        //     return <FruitSelector handlers={{onCreate: selHdl, isTopLevel: false}}/>
        // },
    }}/>
}