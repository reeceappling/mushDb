import {Note} from "@/app/components/formSubcomponents/notes";
import {
    Contamination,
} from "@/app/components/formSubcomponents/contaminations";
import {PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import {ACL, TestAcl} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {PlateSelector} from "@/app/components/plateClient";

export function TestPlateOk(){
    const now = new Date().getTime()
    const testNote = ()=>{
        return {time: new Date().getTime(), note:"TEST_NOTE_TEXT_HERE"}
    }
    const testNotes: Note[] = [testNote(), testNote(), testNote()]
    const aPic: PicWithNotesIncoming = {time: now, notes: [...testNotes], location: "test.jpg"}
    const p: PicWithNotesIncoming[] = [aPic,aPic,aPic]
    const c: Contamination = {time: now, location: "test.jpg", mold:true, bacteria:false, confirmed:true, notes: [...testNotes]}
    return new PlateData({
        _id: "(PLATE ID HERE)",
        agarBatch: "(AGAR BATCH ID)", // TODO: used to be agar?
        creationDate: Date.now()-2000,
        species: "(SPECIES NAME)",
        subspecies: "(SUBSPECIES NAME)",
        innoc: "(Innoc transfer id!)",
        genSpore: 7,
        genFruitOrSpore: 3,
        transfersOut: ["(TRANSFER 1)","(TRANSFER 2)"],
        parentType: "plate",
        parent: "(PARENT ID)",
        pics: [...p],
        contamination: [...[c,c]],
        knownFruitable: true,
        sale: "SALE_ID_HERE",
        disposed: Date.now()+40000,
        mostRecentImage: {...p[0]},
        notes: [...testNotes],
        lastUpdated: 789,
        acl: TestAcl(),
    })
}
export interface PlateData {
    _id: string
    agarBatch?: string
    creationDate: number
    condensationCoverageAtPourTime?: number
    condensationCoverageAtSealTime?: number
    pourCoverage?: number
    wetAtCooledTime?: boolean
    agarOnOutsideAtPourTime?: boolean
    species?: string
    subspecies?: string
    innoc?: string
    genSpore?:  number
    genFruitOrSpore?: number
    transfersOut?: string[]
    parentType?: string
    parent?: string
    contamination?: Contamination[]
    knownFruitable?: boolean
    sale?: string
    disposed?: number
    pics?: PicWithNotesIncoming[]
    mostRecentImage?: PicWithNotesIncoming
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class PlateData {
    // Accept a single object containing the fields
    constructor(init?: Partial<PlateData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "plate"
    }
    public description(): string {
        if(this.species !== undefined){
            return `Plate ${this._id}. Species ${this.species}. ${this.subspecies!==undefined&&`Subspecies ${this.subspecies}`}. Created on ${new Date(this.creationDate).toISOString()}. Last updated on ${new Date(this.lastUpdated).toISOString()}.${this.disposed!==undefined&&` Disposed on ${new Date(this.disposed).toISOString()}`}` // TODO: KF, contams, etc?
        }
        return `Plate ${this._id}. Not innoculated. Created on ${new Date(this.creationDate).toISOString()}.${this.disposed!==undefined&&` Disposed on ${new Date(this.disposed).toISOString()}`}`
    }
}

export function PlateSelectorCloseable(sp: SelectorProps<PlateData>) { // TODO: use
    const doSel = (val?: PlateData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<PlateData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Plate List",
        //createTxt: "Create Bag",// TODO: ???
        lowercase: "plate",
        //creatorInPage: sp.creatorInPage,// TODO: ???
        //createEndpt: "bag",// TODO: ???
        createSelector:(selHdl: (onSelect: PlateData) => void)=>{
            return <PlateSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        // TODO: ok?
        // createCreator:(selHdl: (onSelect: FruitingChamberData) => void)=>{
        //     return <NewFruitingChamberForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        // },
    }}/>
}