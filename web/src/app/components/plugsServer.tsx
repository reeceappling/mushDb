import {Note} from "@/app/components/formSubcomponents/notes";
import {ACL, TestAcl} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {PlugsSelector} from "@/app/components/plugsClient";
import {PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import {Contamination} from "@/app/components/formSubcomponents/contaminations";

export function TestPlugsOk(){
    const testNote = ()=>{
        return {time: new Date().getTime(), note:"TEST_NOTE_TEXT_HERE"}
    }
    const testNotes: Note[] = [testNote(), testNote(), testNote()]
    return new PlugsData({
        _id: "(PLATE ID HERE)",
        dowelTypes: [
            {wood:"wood1",size:1,units:"miles"},
            {wood:"wood2",size:0.5,units:"cm"},
        ],
        creationDate: Date.now()-2000,
        species: "(SPECIES NAME)",
        subspecies: "(SUBSPECIES NAME)",
        innoc: "(Innoc transfer id!)",
        genSpore: 7,
        genFruitOrSpore: 3,
        parentType: "plate",
        parent: "(PARENT ID)",
        knownFruitable: true,
        sales: ["SALE_ID_HERE","SALE_ID_2_HERE"],
        disposed: Date.now()+40000,
        notes: [...testNotes],
        lastUpdated: 789,
        acl: TestAcl(),
    })
}
export interface DowelType {
    wood: string
    size: number
    units: string
}
export interface PlugsData {
    _id: string
    parentType?: string
    parent?: string
    creationDate: number
    dowelTypes: DowelType[]
    genSpore?: number
    genFruitOrSpore?: number
    species?: string
    subspecies?: string
    innoc?: string
    transfersOut?: string[]
    knownFruitable?: boolean
    pcRun?:string
    pics?: PicWithNotesIncoming[]
    contamination?: Contamination[]
    mostRecentImage?: PicWithNotesIncoming
    sales?: string[]
    disposed?: number
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class PlugsData {
    // Accept a single object containing the fields
    constructor(init?: Partial<PlugsData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public entryType(): string {
        return "plugs" // TODO: ensure ok
    }
    public description(): string {
        const firstSent = `Plugs jar ${this._id}`
        // TODO: plugs types?
        const secondSent = this.species !== undefined?`Species ${this.species}${this.subspecies!==undefined&&`. Subspecies ${this.subspecies}`}`:` Not innoculated`
        const lastSent = `Created on ${new Date(this.creationDate).toISOString()}. Last updated on ${new Date(this.lastUpdated).toISOString()}.${this.disposed !== undefined && ` Disposed on ${new Date(this.disposed).toISOString()}`}`
        return `${firstSent}. ${secondSent}. ${lastSent}`
    }
}

export function PlugsSelectorCloseable(sp: SelectorProps<PlugsData>) { // TODO: use
    const doSel = (val?: PlugsData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<PlugsData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Plugs List",
        //createTxt: "Create Bag",// TODO: ???
        lowercase: "plugs",
        //creatorInPage: sp.creatorInPage,// TODO: ???
        //createEndpt: "bag",// TODO: ???
        createSelector:(selHdl: (onSelect: PlugsData) => void)=>{
            return <PlugsSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        // TODO: ok?
        // createCreator:(selHdl: (onSelect: FruitingChamberData) => void)=>{
        //     return <NewFruitingChamberForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        // },
    }}/>
}