import {Note} from "@/app/components/formSubcomponents/notes";
import {ACL} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {ChannelTextNewAgarBatch} from "@/app/components/agarBatchServer";
import {PlateSelector} from "@/app/components/plateClient";
import {PlateData} from "@/app/components/plateServer";
import {PlugsSelector} from "@/app/components/plugsClient";

export function TestPlugsOk(){
    const testNote = ()=>{
        return {time: new Date().getTime(), note:"TEST_NOTE_TEXT_HERE"}
    }
    const testNotes: Note[] = [testNote(), testNote(), testNote()]
    const a: PlugsJar = {
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
    }
    return a
}
export interface DowelType {
    wood: string
    size: number
    units: string
}
export interface PlugsJar {
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
    sales?: string[]
    disposed?: number
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}

export function PlugsSelectorCloseable(sp: SelectorProps<PlugsJar>) { // TODO: use
    const doSel = (val?: PlugsJar):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<PlugsJar> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        msgTxt: ChannelTextNewAgarBatch, // TODO: ???
        closeTxt: "Close Plugs List",
        //createTxt: "Create Bag",// TODO: ???
        lowercase: "plugs",
        //creatorInPage: sp.creatorInPage,// TODO: ???
        //createEndpt: "bag",// TODO: ???
        getId: (v: PlugsJar) => v._id,
        createSelector:(selHdl: (onSelect: PlugsJar) => void)=>{
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